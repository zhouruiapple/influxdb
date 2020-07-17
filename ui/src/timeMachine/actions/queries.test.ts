import {mocked} from 'ts-jest/utils'

import {executeQueries} from 'src/timeMachine/actions/queries'
import configureStore from 'src/store/configureStore'
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

describe('executing queries', () => {
  describe('executing an empty list of queries', () => {
    const dispatchExecuteQueries = executeQueries()
    const store = configureStore(localState)
    const {dispatch} = store
    const dispatchMock = jest.spyOn(store, 'dispatch')

    dispatchExecuteQueries(dispatchMock, store.getState)
    const [
      noQueriesDispatch,
      loadingDispatch,
      hydrateVars,
    ] = dispatchMock.mock.calls

    it('dispatches setQueryResults to Done when there are no queries', () => {
      expect(dispatchMock.mock.calls.length).toBe(3)

      expect(noQueriesDispatch[0].type).toBe('SET_QUERY_RESULTS')
      expect(noQueriesDispatch[0].payload.status).toBe('Done')
      expect(noQueriesDispatch[0].payload.files).toEqual([])
    })

    it('does something with loadingDispatch next', () => {
      console.error('loadingDispatch: ', loadingDispatch)
    })
  })
})
