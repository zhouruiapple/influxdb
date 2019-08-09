import React, {FC} from 'react'
import {Plot, Table, Config} from '@influxdata/giraffe'

import {VIS_THEME} from 'src/shared/constants'

interface Props {
  table: Table
}

const LogsHistogram: FC<Props> = ({table}) => {
  const config: Config = {
    ...VIS_THEME,
    gridOpacity: 0.3,
    axisOpacity: 0.5,
    table,
    layers: [
      {
        type: 'histogram',
        x: 'time',
        fill: ['severity'],
        binCount: 50,
      },
    ],
  }

  return (
    <div className="logs-histogram">{table && <Plot config={config} />}</div>
  )
}

export default LogsHistogram
