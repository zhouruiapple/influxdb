import {Table, Config, isNumeric} from '@influxdata/vis'
import {ToMinardTableResult} from 'src/shared/utils/toMinardTable'
import {produce} from 'immer'

export const getNumericColumns = (table: Table): string[] => {
  const numericColumns = Object.entries(table.columns)
    .filter(([__, {type}]) => isNumeric(type))
    .map(([name]) => name)

  return numericColumns
}

const INVALID_GROUP_COLUMNS = new Set(['_value', '_start', '_stop', '_time'])

export const getGroupableColumns = (table: Table): string[] => {
  const groupableColumns = Object.keys(table.columns).filter(
    name => !INVALID_GROUP_COLUMNS.has(name)
  )

  return groupableColumns
}

export const resolveNumericMapping = (
  validColumns: string[],
  preferredColumn: string
): string => {
  if (preferredColumn && validColumns.includes(preferredColumn)) {
    return preferredColumn
  }

  if (validColumns.includes('_value')) {
    return '_value'
  }

  if (validColumns.length) {
    return validColumns[0]
  }

  return null
}

export const resolveGroupMapping = (
  validColumns: string[],
  preferredColumns: string[]
): string[] => {
  if (
    preferredColumns &&
    preferredColumns.every(col => validColumns.includes(col))
  ) {
    return preferredColumns
  }

  // TODO: Fall back to the union of all Flux group keys
  return []
}

const AESTHETIC_TYPES = {
  x: 'numeric',
  y: 'numeric',
  fill: 'group',
  symbol: 'group',
}

const DEFAULT_NUMERIC_MAPPINGS = {
  x: '_time',
  y: '_value',
}

export const resolveMappings = (
  config: Partial<Config>,
  tableResult: ToMinardTableResult
): Partial<Config> => {
  const table = tableResult.table
  const numericColumns = getNumericColumns(table)
  const groupableColumns = getGroupableColumns(table)
  const defaultGroupColumns = tableResult.defaultGroupColumns.filter(
    k => !INVALID_GROUP_COLUMNS.has(k)
  )

  return produce(config, draftConfig => {
    for (const layer of draftConfig.layers) {
      for (const [aes, aesType] of Object.entries(AESTHETIC_TYPES)) {
        if (!layer.hasOwnProperty(aes)) {
          continue
        } else if (aesType === 'numeric') {
          layer[aes] = resolveNumericMapping(
            numericColumns,
            layer[aes] || DEFAULT_NUMERIC_MAPPINGS[aes]
          )
        } else if (aesType === 'group') {
          layer[aes] = resolveGroupMapping(
            groupableColumns,
            layer[aes] || defaultGroupColumns
          )
        }
      }
    }
  })
}
