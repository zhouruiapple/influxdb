// Libraries
import {useMemo, FunctionComponent} from 'react'
import {connect} from 'react-redux'
import {Config} from '@influxdata/vis'

// Utils
import {getVisTable} from 'src/timeMachine/selectors'
import {resolveMappings} from 'src/shared/utils/vis'

// Constants
import {VIS_DEFAULTS} from 'src/shared/constants'

// Types
import {AppState} from 'src/types'
import {ToMinardTableResult} from 'src/shared/utils/toMinardTable'

interface StateProps {
  tableResult: ToMinardTableResult
}

interface OwnProps {
  config: Config
  children: (config: Config) => JSX.Element
}

type Props = StateProps & OwnProps

const VisTransform: FunctionComponent<Props> = ({
  config,
  tableResult,
  children,
}) => {
  // TODO: One way binding for x/y domains
  const resolvedConfig = useMemo(
    () => ({
      ...VIS_DEFAULTS,
      ...resolveMappings(config, tableResult),
      table: tableResult.table,
    }),
    [tableResult, config, VIS_DEFAULTS]
  )

  return children(resolvedConfig)
}

const mstp = (state: AppState) => {
  const tableResult = getVisTable(state)

  return {tableResult}
}

export default connect<StateProps, {}, OwnProps>(mstp)(VisTransform)
