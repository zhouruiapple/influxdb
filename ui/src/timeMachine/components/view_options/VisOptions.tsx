// Libraries
import React, {useMemo, FunctionComponent} from 'react'
import {connect} from 'react-redux'
import {Config, LayerConfig} from '@influxdata/vis'
import {produce} from 'immer'
import {
  Grid,
  Button,
  ComponentColor,
  ComponentSize,
  IconFont,
} from '@influxdata/clockface'

// Components
import VisBaseOptions from 'src/timeMachine/components/view_options/VisBaseOptions'
import VisLayerOptions from 'src/timeMachine/components/view_options/VisLayerOptions'
import VisLineLayerOptions from 'src/timeMachine/components/view_options/VisLineLayerOptions'

// Actions
import {setConfig} from 'src/timeMachine/actions'

// Utils
import {getActiveTimeMachine} from 'src/timeMachine/selectors'
import {getVisTable} from 'src/timeMachine/selectors'
import {resolveMappings} from 'src/shared/utils/vis'

// Types
import {AppState, WorkingView, VisView} from 'src/types'
import {ToMinardTableResult} from 'src/shared/utils/toMinardTable'

interface StateProps {
  config: Partial<Config>
  tableResult: ToMinardTableResult
}

interface DispatchProps {
  onSetConfig: (config: Partial<Config>) => void
}

type Props = StateProps & DispatchProps

const VisOptions: FunctionComponent<Props> = ({
  config,
  onSetConfig,
  tableResult,
}) => {
  const resolvedConfig = useMemo(() => resolveMappings(config, tableResult), [
    config,
    tableResult.table,
  ])

  const onSetLayer = i => (updatedLayer: LayerConfig) => {
    onSetConfig(
      produce(resolvedConfig, draftConfig => {
        draftConfig.layers[i] = updatedLayer
      })
    )
  }

  return (
    <Grid.Column>
      <h5 className="view-options--header">Layers</h5>
      <Button
        text="Add Layer"
        color={ComponentColor.Primary}
        icon={IconFont.Plus}
        size={ComponentSize.ExtraSmall}
      />
      {resolvedConfig.layers.map((layer, i) => {
        if (layer.type === 'line') {
          return (
            <VisLayerOptions layer={layer} key={`${i}-${layer.type}`}>
              <VisLineLayerOptions layer={layer} onSetLayer={onSetLayer(i)} />
            </VisLayerOptions>
          )
        }

        return null
      })}
      <h5 className="view-options--header">Base Options</h5>
      <VisBaseOptions config={resolvedConfig} onSetConfig={onSetConfig} />
    </Grid.Column>
  )
}

const mstp = (state: AppState) => {
  const view = getActiveTimeMachine(state).view as WorkingView<VisView>
  const config = view.properties.config
  const tableResult = getVisTable(state)

  return {config, tableResult}
}

const mdtp = {
  onSetConfig: setConfig,
}

export default connect<StateProps, DispatchProps>(
  mstp,
  mdtp
)(VisOptions)
