// Libraries
import React, {FunctionComponent} from 'react'
import {Config, Table} from '@influxdata/giraffe'

// Components
// import EmptyGraphMessage from 'src/shared/components/EmptyGraphMessage'

// Utils
import {
  useVisXDomainSettings,
  useVisYDomainSettings,
} from 'src/shared/utils/useVisDomainSettings'
import {
  getFormatter,
  defaultXColumn,
  mosaicYcolumn,
  mosaicFillColumn,
} from 'src/shared/utils/vis'

// Constants
import {VIS_THEME, VIS_THEME_LIGHT} from 'src/shared/constants'
import {DEFAULT_LINE_COLORS} from 'src/shared/constants/graphColorPalettes'
// import {INVALID_DATA_COPY} from 'src/shared/copy/cell'

// Types
import {MosaicViewProperties, TimeZone, TimeRange, Theme} from 'src/types'

interface Props {
  children: (config: Config) => JSX.Element
  fluxGroupKeyUnion?: string[]
  timeRange: TimeRange | null
  table: Table
  timeZone: TimeZone
  viewProperties: MosaicViewProperties
  theme?: Theme
}

const MosaicPlot: FunctionComponent<Props> = ({
  children,
  timeRange,
  timeZone,
  table,
  viewProperties: {
    xAxisLabel,
    yAxisLabel,
    xPrefix,
    xSuffix,
    yPrefix,
    ySuffix,
    fillColumn: storedFill,
    colors,
    xDomain: storedXDomain,
    yDomain: storedYDomain,
    xColumn: storedXColumn,
    yColumn: storedYColumn,
    timeFormat,
  },
  theme,
}) => {
  const fillColumn = storedFill || mosaicFillColumn(table)
  console.log('fillColumn mosaicPlot', fillColumn)

  console.log('table', table)
  console.log('timeRange', timeRange)
  console.log('storedYColumn', storedYColumn)
  const xColumn = storedXColumn || defaultXColumn(table)
  const yColumn = storedYColumn || mosaicYcolumn(table) //, 'taskID'
  //const stringFillColumn = storedStringFillColumn || defaultStringFillColumn(table, '_value')

  const columnKeys = table.columnKeys

  const [xDomain, onSetXDomain, onResetXDomain] = useVisXDomainSettings(
    storedXDomain,
    table.getColumn(xColumn, 'number'),
    timeRange
  )

  console.log('storedXDomain', storedXDomain)
  console.log('yColumn', yColumn)
  const [yDomain, onSetYDomain, onResetYDomain] = useVisYDomainSettings(
    storedYDomain,
    table.getColumn(yColumn, 'string')
  )

  console.log('yDomain', yDomain)

  // const isValidView =
  //   xColumn &&
  //   columnKeys.includes(xColumn) &&
  //   yColumn &&
  //   columnKeys.includes(yColumn) &&
  //   fillColumns.every(col => columnKeys.includes(col))

  // if (!isValidView) {
  //   return <EmptyGraphMessage message={INVALID_DATA_COPY} />
  // }

  const colorHexes =
    colors && colors.length ? colors : DEFAULT_LINE_COLORS.map(c => c.hex)

  const xFormatter = getFormatter(table.getColumnType(xColumn), {
    prefix: xPrefix,
    suffix: xSuffix,
    timeZone,
    timeFormat,
  })

  const yFormatter = getFormatter(table.getColumnType(yColumn), {
    prefix: yPrefix,
    suffix: ySuffix,
    timeZone,
    timeFormat,
  })

  const currentTheme = theme === 'light' ? VIS_THEME_LIGHT : VIS_THEME

  const config: Config = {
    ...currentTheme,
    table,
    xAxisLabel,
    yAxisLabel,
    xDomain,
    onSetXDomain,
    onResetXDomain,
    yDomain,
    onSetYDomain,
    onResetYDomain,
    valueFormatters: {
      [xColumn]: xFormatter,
      [yColumn]: yFormatter,
    },
    layers: [
      {
        type: 'mosaic',
        x: xColumn,
        y: yColumn,
        colors: colorHexes,
        fill: fillColumn,
      },
    ],
  }
  // console.log('children(config)', children(config))
  return children(config)
}
// console.log('mosaic plot', MosaicPlot)
export default MosaicPlot
