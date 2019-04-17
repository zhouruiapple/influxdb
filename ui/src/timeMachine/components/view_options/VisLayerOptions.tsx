// Libraries
import React, {useState, FunctionComponent} from 'react'
import {LayerConfig} from '@influxdata/vis'
import {Icon, IconFont} from '@influxdata/clockface'
import classnames from 'classnames'

const LAYER_DISPLAY_NAMES = {
  line: 'Line',
  histogram: 'Histogram',
  heatmap: 'Heatmap',
  scatterplot: 'Scatterplot',
}

interface Props {
  layer: LayerConfig
}

const VisLayerOptions: FunctionComponent<Props> = ({layer, children}) => {
  const [isActive, setIsActive] = useState(false)

  const className = classnames('vis-layer-options', {
    active: isActive,
  })

  return (
    <div className={className}>
      <div
        className="vis-layer-options--header"
        onClick={() => setIsActive(!isActive)}
      >
        <Icon glyph={isActive ? IconFont.CaretDown : IconFont.CaretRight} />
        {LAYER_DISPLAY_NAMES[layer.type] || 'Layer'}
      </div>
      {isActive && (
        <div className="vis-layer-options--children">{children}</div>
      )}
    </div>
  )
}

export default VisLayerOptions
