// Libraries
import memoizeOne from 'memoize-one'
import {get, flatMap} from 'lodash'

// Utils
import {parseResponse} from 'src/shared/parsing/flux/response'
import {
  toMinardTable,
  ToMinardTableResult,
} from 'src/shared/utils/toMinardTable'
import {
  getNumericColumns as getNumericColumnsFn,
  getGroupableColumns as getGroupableColumnsFn,
  resolveMappings,
  resolveNumericMapping,
  resolveGroupMapping,
} from 'src/shared/utils/vis'

// Types
import {
  FluxTable,
  QueryView,
  AppState,
  DashboardDraftQuery,
  ViewType,
} from 'src/types'

export const getActiveTimeMachine = (state: AppState) => {
  const {activeTimeMachineID, timeMachines} = state.timeMachines
  const timeMachine = timeMachines[activeTimeMachineID]

  return timeMachine
}

export const getActiveQuery = (state: AppState): DashboardDraftQuery => {
  const {draftQueries, activeQueryIndex} = getActiveTimeMachine(state)

  return draftQueries[activeQueryIndex]
}

const getTablesMemoized = memoizeOne(
  (files: string[]): FluxTable[] => (files ? flatMap(files, parseResponse) : [])
)

export const getTables = (state: AppState): FluxTable[] =>
  getTablesMemoized(getActiveTimeMachine(state).queryResults.files)

const getVisTableMemoized = memoizeOne(toMinardTable)

export const getVisTable = (state: AppState): ToMinardTableResult => {
  const fluxTables = getTables(state)
  const result = getVisTableMemoized(fluxTables)

  return result
}

const getNumericColumnsMemoized = memoizeOne(getNumericColumnsFn)

export const getNumericColumns = (state: AppState): string[] => {
  const table = getVisTable(state).table

  return getNumericColumnsMemoized(table)
}

const getGroupableColumnsMemoized = memoizeOne(getGroupableColumnsFn)

export const getGroupableColumns = (state: AppState): string[] => {
  const table = getVisTable(state).table

  return getGroupableColumnsMemoized(table)
}

const getXColumnSelectionMemoized = memoizeOne(resolveNumericMapping)

export const getXColumnSelection = (state: AppState): string => {
  const validXColumns = getNumericColumns(state)
  const preference = get(getActiveTimeMachine(state), 'view.properties.xColumn')

  return getXColumnSelectionMemoized(validXColumns, preference)
}

const getFillColumnsSelectionMemoized = memoizeOne(resolveGroupMapping)

export const getFillColumnsSelection = (state: AppState): string[] => {
  const validFillColumns = getGroupableColumns(state)
  const preference = get(
    getActiveTimeMachine(state),
    'view.properties.fillColumns'
  )

  return getFillColumnsSelectionMemoized(validFillColumns, preference)
}

export const getSaveableView = (state: AppState): QueryView & {id?: string} => {
  const {view, draftQueries} = getActiveTimeMachine(state)

  let saveableView: QueryView & {id?: string} = {
    ...view,
    properties: {
      ...view.properties,
      queries: draftQueries,
    },
  }

  if (saveableView.properties.type === ViewType.Histogram) {
    saveableView = {
      ...saveableView,
      properties: {
        ...saveableView.properties,
        xColumn: getXColumnSelection(state),
        fillColumns: getFillColumnsSelection(state),
      },
    }
  } else if (saveableView.properties.type === ViewType.Vis) {
    saveableView = {
      ...saveableView,
      properties: {
        ...saveableView.properties,
        config: resolveMappings(
          saveableView.properties.config,
          getVisTable(state)
        ),
      },
    }
  }

  return saveableView
}
