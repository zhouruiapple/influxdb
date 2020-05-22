// Libraries
import {produce} from 'immer'

// Types
import {Snippit, RemoteDataState, ResourceState, ResourceType} from 'src/types'
import {
  ADD_SNIPPIT,
  SET_SNIPPITS,
  Action,
  EDIT_SNIPPIT,
  REMOVE_SNIPPIT,
} from 'src/snippits/actions/creators'

// Utils
import {
  setResource,
  addResource,
  removeResource,
  editResource,
} from 'src/resources/reducers/helpers'

const {Snippits} = ResourceType
type SnippitsState = ResourceState['snippits']

const initialState = (): SnippitsState => ({
  status: RemoteDataState.NotStarted,
  byID: {},
  allIDs: [],
})

export const snippitsReducer = (
  state: SnippitsState = initialState(),
  action: Action
): SnippitsState =>
  produce(state, draftState => {
    switch (action.type) {
      case SET_SNIPPITS: {
        setResource<Snippit>(draftState, action, Snippits)

        return
      

      case ADD_SNIPPIT: {
        addResource<Snippit>(draftState, action, Snippits)

        return
      }

      case EDIT_SNIPPIT: {
        editResource<Snippit>(draftState, action, Snippits)

        return
      }

      case REMOVE_SNIPPIT: {
        removeResource<Snippit>(draftState, action)

        return
      }
    }
  })
