// Libraries
import React, {FunctionComponent} from 'react'
import {Config} from '@influxdata/vis'
import {Form, Input} from '@influxdata/clockface'
import AutoDomainInput from 'src/shared/components/AutoDomainInput'

interface Props {
  config: Partial<Config>
  onSetConfig: (config: Partial<Config>) => void
}

const VisBaseOptions: FunctionComponent<Props> = ({config, onSetConfig}) => {
  return (
    <>
      <Form.Element label="X Axis Label">
        <Input
          value={config.xAxisLabel}
          onChange={e => onSetConfig({...config, xAxisLabel: e.target.value})}
        />
      </Form.Element>
      <Form.Element label="Y Axis Label">
        <Input
          value={config.yAxisLabel}
          onChange={e => onSetConfig({...config, yAxisLabel: e.target.value})}
        />
      </Form.Element>
      <AutoDomainInput
        domain={config.xDomain}
        onSetDomain={d => onSetConfig({...config, xDomain: d})}
        label="Set X Axis Domain"
      />
      <AutoDomainInput
        domain={config.yDomain}
        onSetDomain={d => onSetConfig({...config, yDomain: d})}
        label="Set Y Axis Domain"
      />
    </>
  )
}

export default VisBaseOptions
