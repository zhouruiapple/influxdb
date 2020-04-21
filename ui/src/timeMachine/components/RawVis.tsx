// Libraries
import React, {FC, useState} from 'react'
import {connect} from 'react-redux'
import {FromFluxResult} from '@influxdata/giraffe'
import {AutoSizer} from 'react-virtualized'
import classnames from 'classnames'

// Components
import EmptyQueryView, {ErrorFormat} from 'src/shared/components/EmptyQueryView'
import RawFluxDataTable from 'src/timeMachine/components/RawFluxDataTable'
import ErrorBoundary from 'src/shared/components/ErrorBoundary'
import CSVExportButton from 'src/shared/components/CSVExportButton'
import {Icon, IconFont} from '@influxdata/clockface'

// Utils
import {getActiveTimeMachine} from 'src/timeMachine/selectors'
import {checkResultsLength} from 'src/shared/utils/vis'
import {
  getVisTable,
  getXColumnSelection,
  getYColumnSelection,
  getFillColumnsSelection,
  getSymbolColumnsSelection,
  getIsInCheckOverlay,
} from 'src/timeMachine/selectors'
import {getTimeRange} from 'src/dashboards/selectors'

// Types
import {
  RemoteDataState,
  AppState,
  QueryViewProperties,
  TimeZone,
  TimeRange,
  StatusRow,
  CheckType,
  Threshold,
} from 'src/types'

// Selectors
import {getActiveTimeRange} from 'src/timeMachine/selectors/index'

interface StateProps {
  timeRange: TimeRange | null
  loading: RemoteDataState
  errorMessage: string
  files: string[]
  viewProperties: QueryViewProperties
  isInitialFetch: boolean
  giraffeResult: FromFluxResult
  xColumn: string
  yColumn: string
  checkType: CheckType
  checkThresholds: Threshold[]
  fillColumns: string[]
  symbolColumns: string[]
  timeZone: TimeZone
  statuses: StatusRow[][]
  isInCheckOverlay: boolean
}

type Props = StateProps

const TimeMachineVis: FC<Props> = ({
  loading,
  errorMessage,
  isInitialFetch,
  files,
  viewProperties,
  giraffeResult,
  isInCheckOverlay,
}) => {
  const [blockMode, setBlockMode] = useState<'expanded' | 'collapsed'>('expanded')

  const timeMachineBlockClass = classnames('tm-block', {[`tm-block__${blockMode}`]: blockMode})

  const handleToggleClick = (): void => {
    const newBlockMode = blockMode === 'expanded' ? 'collapsed' : 'expanded'
    setBlockMode(newBlockMode)
  }

  return (
    <div className={timeMachineBlockClass}>
      <div className="tm-block--header">
        <div className="tm-block--header-left">
          <button className="tm-block--toggle" onClick={handleToggleClick}>
            <Icon glyph={IconFont.CaretRight} className="tm-block--toggle-icon" />
          </button>
          <div className="tm-block--title">Results</div>
        </div>
        <div className="tm-block--header-right">
          {!isInCheckOverlay && <CSVExportButton />}
        </div>
      </div>
      <div className="tm-block--contents">
        <ErrorBoundary>
          <EmptyQueryView
            loading={loading}
            errorFormat={ErrorFormat.Scroll}
            errorMessage={errorMessage}
            isInitialFetch={isInitialFetch}
            queries={viewProperties.queries}
            hasResults={checkResultsLength(giraffeResult)}
          >
            <AutoSizer>
              {({width, height}) =>
                width &&
                height && (
                  <RawFluxDataTable
                    files={files}
                    width={width}
                    height={height}
                  />
                )
              }
            </AutoSizer>
          </EmptyQueryView>
        </ErrorBoundary>
      </div>
    </div>
  )
}

const mstp = (state: AppState): StateProps => {
  const activeTimeMachine = getActiveTimeMachine(state)
  const {
    view: {properties: viewProperties},
    queryResults: {
      status: loading,
      errorMessage,
      isInitialFetch,
      files,
      statuses,
    },
  } = activeTimeMachine
  const timeRange = getTimeRange(state)
  const {
    alertBuilder: {type: checkType, thresholds: checkThresholds},
  } = state

  const giraffeResult = getVisTable(state)
  const xColumn = getXColumnSelection(state)
  const yColumn = getYColumnSelection(state)
  const fillColumns = getFillColumnsSelection(state)
  const symbolColumns = getSymbolColumnsSelection(state)

  const timeZone = state.app.persisted.timeZone

  return {
    loading,
    checkType,
    checkThresholds,
    errorMessage,
    isInitialFetch,
    files,
    viewProperties,
    giraffeResult,
    xColumn,
    yColumn,
    fillColumns,
    symbolColumns,
    timeZone,
    timeRange: getActiveTimeRange(timeRange, viewProperties.queries),
    isInCheckOverlay: getIsInCheckOverlay(state),
    statuses,
  }
}

export default connect<StateProps>(mstp)(TimeMachineVis)
