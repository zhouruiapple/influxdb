// Libraries
import {produce} from 'immer'

// Types
import {RemoteDataState} from 'src/types'
import {Action} from 'src/globalSearch/actions'

const initialState = (): GlobalSearchState => ({
  status: RemoteDataState.NotStarted,
  list: [],
})

export interface GlobalSearchState {
  status: RemoteDataState
  list: any[]
}

export const globalSearchReducer = (
  state: GlobalSearchState = initialState(),
  action: Action
): GlobalSearchState =>
  produce(state, draftState => {
    switch (action.type) {
      case 'SET_GLOBAL_SEARCH': {
        const {status, list} = action.payload

        draftState.status = status

        if (list) {
          draftState.list = list
        }

        return
      }
    }
  })
