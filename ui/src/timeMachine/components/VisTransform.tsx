// Libraries
import {FunctionComponent} from 'react'
import {connect} from 'react-redux'
import {Table, Config} from '@influxdata/vis'

// Utils
import {
  getVisTable,
  getNumericColumns,
  getGroupableColumns,
} from 'src/timeMachine/selectors'

// Constants
import {DEFAULT_LINE_COLORS} from 'src/shared/constants/graphColorPalettes'
import {VIS_DEFAULTS} from 'src/shared/constants'

// Types
import {AppState} from 'src/types'

interface StateProps {
  table: Table
  numericColumns: string[]
  groupColumns: string[]
}

interface OwnProps {
  config: Config
  children: (config: Config) => JSX.Element
}

type Props = StateProps & OwnProps

const VisTransform: FunctionComponent<Props> = ({config, table, children}) => {
  const fullConfig: Config = {
    ...VIS_DEFAULTS,
    ...config,
    table,
    yTickFormatter: t => `${Math.round(t)}%`,
    layers: [
      {
        type: 'line',
        x: '_time',
        y: '_value',
        fill: ['_field', 'cpu'],
        colors: DEFAULT_LINE_COLORS.map(c => c.hex),
        interpolation: 'monotoneX',
      },
    ],
  }

  return children(fullConfig)
}

const mstp = (state: AppState) => {
  const table = getVisTable(state)
  const numericColumns = getNumericColumns(state)
  const groupColumns = getGroupableColumns(state)

  return {table, numericColumns, groupColumns}
}

export default connect<StateProps, {}, OwnProps>(mstp)(VisTransform)
