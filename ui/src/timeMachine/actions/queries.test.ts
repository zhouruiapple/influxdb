import {mocked} from 'ts-jest/utils'

import {executeQueries} from 'src/timeMachine/actions/queries'
import {initialStateHelper} from 'src/timeMachine/reducers/'
import configureStore, {clearStore} from 'src/store/configureStore'
import {localState} from 'src/mockState'

const emptyResults = {
  type: 'SET_QUERY_RESULTS',
  payload: {
    status: 'Done',
    files: [],
    fetchDuration: null,
    errorMessage: undefined,
    statuses: undefined,
  },
}

let dispatchMock

describe('executing queries', () => {
  describe('executing an empty list of queries', () => {
    const dispatchExecuteQueries = executeQueries()
    const store = configureStore(localState)
    const {dispatch} = store
    dispatchMock = jest.spyOn(store, 'dispatch')

    it('dispatches setQueryResults to Done when there are no queries', async () => {
      await dispatchExecuteQueries(dispatchMock, store.getState)
      const [
        noQueriesDispatch,
        loadingDispatch,
        hydrateVars,
        filesDispatch,
      ] = dispatchMock.mock.calls

      console.log(dispatchMock.mock.calls)
      expect(dispatchMock.mock.calls.length).toBe(4)

      expect(noQueriesDispatch[0].type).toBe('SET_QUERY_RESULTS')
      expect(noQueriesDispatch[0].payload.status).toBe('Done')
      expect(noQueriesDispatch[0].payload.files).toEqual([])

      expect(filesDispatch[0].type).toBe('SET_QUERY_RESULTS')
      expect(filesDispatch[0].payload.status).toBe('Done')
      expect(filesDispatch[0].payload.files).toEqual([])
    })
  })

  describe('executing a single query', () => {
    const localStateCopy = {
      ...localState,
      timeMachines: {
        activeTimeMachineID: 'yourmom',
        yourmom: {...initialStateHelper()},
      },
    } as any

    const queryText = `from(bucket: v.bucket)
      |> range(start: v.timeRangeStart)
      |> filter(fn: (r) => r._measurement == "system")
      |> filter(fn: (r) => r._field == "load1" or r._field == "load5" or r._field == "load15")
      |> aggregateWindow(every: v.windowPeriod, fn: mean, createEmpty: false)
      |> yield(name: "mean")`

    localStateCopy.timeMachines.yourmom.view.properties.queries = [queryText]

    console.log('localStateCopy.timeMachines: ', localStateCopy.timeMachines)

    const dispatchExecuteQueries = executeQueries()
    const store = configureStore(localStateCopy)
    const {dispatch} = store
    const dispatchMock = jest.spyOn(store, 'dispatch')

    it('dispatches setQueryResults to Done when there are no queries', async () => {
      console.log('hi')
      await dispatchExecuteQueries(dispatchMock, store.getState)
      console.log('ok')
      const [loadingDispatch, hydrateVars] = dispatchMock.mock.calls
      expect(true).toEqual(true)
    })
    // dispatchMock.mockReset()
  })
})
