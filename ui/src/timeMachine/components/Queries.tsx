// Libraries
import React, {FC, useState} from 'react'
import {connect} from 'react-redux'
import classnames from 'classnames'

// Components
import TimeMachineFluxEditor from 'src/timeMachine/components/TimeMachineFluxEditor'
import TimeMachineQueriesSwitcher from 'src/timeMachine/components/QueriesSwitcher'
import TimeMachineRefreshDropdown from 'src/timeMachine/components/RefreshDropdown'
import TimeRangeDropdown from 'src/shared/components/TimeRangeDropdown'
import TimeMachineQueryBuilder from 'src/timeMachine/components/QueryBuilder'
import SubmitQueryButton from 'src/timeMachine/components/SubmitQueryButton'
import EditorShortcutsToolTip from 'src/timeMachine/components/EditorShortcutsTooltip'
import {
  Icon,
  IconFont,
} from '@influxdata/clockface'

// Actions
import {setAutoRefresh} from 'src/timeMachine/actions'
import {setTimeRange} from 'src/timeMachine/actions'

// Utils
import {
  getActiveTimeMachine,
  getIsInCheckOverlay,
  getActiveQuery,
} from 'src/timeMachine/selectors'
import {getTimeRange} from 'src/dashboards/selectors'

// Types
import {
  AppState,
  DashboardQuery,
  TimeRange,
  AutoRefresh,
  AutoRefreshStatus,
} from 'src/types'

interface StateProps {
  activeQuery: DashboardQuery
  timeRange: TimeRange
  autoRefresh: AutoRefresh
  isInCheckOverlay: boolean
}

interface DispatchProps {
  onSetTimeRange: typeof setTimeRange
  onSetAutoRefresh: typeof setAutoRefresh
}

type Props = StateProps & DispatchProps

const TimeMachineQueries: FC<Props> = ({activeQuery, timeRange, autoRefresh, isInCheckOverlay, onSetAutoRefresh, onSetTimeRange}) => {
  const [blockMode, setBlockMode] = useState<'expanded' | 'collapsed'>('expanded')

  const timeMachineBlockClass = classnames('tm-block', {[`tm-block__${blockMode}`]: blockMode})

  const handleToggleClick = (): void => {
    const newBlockMode = blockMode === 'expanded' ? 'collapsed' : 'expanded'
    setBlockMode(newBlockMode)
  }
  
  const handleSetTimeRange = (timeRange: TimeRange) => {
    onSetTimeRange(timeRange)

    if (timeRange.type === 'custom') {
      onSetAutoRefresh({...autoRefresh, status: AutoRefreshStatus.Disabled})
      return
    }

    if (autoRefresh.status === AutoRefreshStatus.Disabled) {
      if (autoRefresh.interval === 0) {
        onSetAutoRefresh({...autoRefresh, status: AutoRefreshStatus.Paused})
        return
      }

      onSetAutoRefresh({...autoRefresh, status: AutoRefreshStatus.Active})
    }
  }

  let queryEditor
  let scriptHelpTooltip

  if (activeQuery.editMode === 'builder') {
    queryEditor = <TimeMachineQueryBuilder />
  } else if (activeQuery.editMode === 'advanced') {
    scriptHelpTooltip = <EditorShortcutsToolTip />
    queryEditor = <TimeMachineFluxEditor />
  }

  let timeControls
  let modeSwitcher

  if (!isInCheckOverlay) {
    timeControls = (
      <>
        <TimeMachineRefreshDropdown />
        <TimeRangeDropdown
          timeRange={timeRange}
          onSetTimeRange={handleSetTimeRange}
        />
      </>
    )

    modeSwitcher = <TimeMachineQueriesSwitcher />
  }

  return (
    <div className={timeMachineBlockClass}>
      <div className="tm-block--header">
        <div className="tm-block--header-left">
          <button className="tm-block--toggle" onClick={handleToggleClick}>
            <Icon glyph={IconFont.CaretRight} className="tm-block--toggle-icon" />
          </button>
          <div className="tm-block--title">Query</div>
          {modeSwitcher}
        </div>
        <div className="tm-block--header-right">
          {scriptHelpTooltip}
          {timeControls}
          <SubmitQueryButton />
        </div>
      </div>
      <div className="tm-block--contents">
        {queryEditor}
      </div>
    </div>
  )
}

const mstp = (state: AppState) => {
  const timeRange = getTimeRange(state)
  const {autoRefresh} = getActiveTimeMachine(state)

  const activeQuery = getActiveQuery(state)

  return {
    timeRange,
    activeQuery,
    autoRefresh,
    isInCheckOverlay: getIsInCheckOverlay(state),
  }
}

const mdtp = {
  onSetTimeRange: setTimeRange,
  onSetAutoRefresh: setAutoRefresh,
}

export default connect<StateProps, DispatchProps>(
  mstp,
  mdtp
)(TimeMachineQueries)
