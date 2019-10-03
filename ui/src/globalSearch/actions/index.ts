// API
// import {client} from 'src/utils/api'

// Types
import {RemoteDataState} from 'src/types'
import {Dispatch} from 'redux-thunk'

// Actions
import {notify} from 'src/shared/actions/notifications'
import {getGlobalSearchFailed} from 'src/shared/copy/notifications'
import {GetState} from 'src/types'
import Axios from 'axios'

export type Action = SetGlobalSearch

interface SetGlobalSearch {
  type: 'SET_GLOBAL_SEARCH'
  payload: {
    status: RemoteDataState
    list: any[]
  }
}

export const setGlobalSearch = (
  status: RemoteDataState,
  list?: any[]
): SetGlobalSearch => ({
  type: 'SET_GLOBAL_SEARCH',
  payload: {status, list},
})

export const getGlobalSearch = (docType: string) => async (
  dispatch: Dispatch<Action>,
  getState: GetState
) => {
  try {
    const {
      orgs: {org},
    } = getState()
    dispatch(setGlobalSearch(RemoteDataState.Loading))

    // const search = await client.search.getAll(org.id)
    console.log('docType', docType)

    let searchResults = ''

    Axios.get('search').then((res: any) => {
      searchResults = res.data
    })

    console.log('response', searchResults)

    const mockResults = [
      {
        IndexType: 'bucket',
        ID: '0000000000000003',
        OrgID: org.id,
        Type: 'user',
        Name: 'bucket 1',
        Description: 'description of bucket 1',
        RetentionPeriod: '1000h0m0s',
        RetentionPolicyName: 'retention name',
        CreatedAt: {
          time: {
            Time: {
              wall: 0x12e42a98,
              ext: 63705733822,
              loc: `(*time.Location)(nil)`,
            },
          },
        },
        UpdatedAt: {
          time: {
            Time: {
              wall: 0x12e42a98,
              ext: 63705733822,
              loc: `(*time.Location)(nil)`,
            },
          },
        },
      },
    ]

    dispatch(setGlobalSearch(RemoteDataState.Done, mockResults))
  } catch (e) {
    console.error(e)
    dispatch(setGlobalSearch(RemoteDataState.Error))
    dispatch(notify(getGlobalSearchFailed()))
  }
}

// /search
//q=docType
