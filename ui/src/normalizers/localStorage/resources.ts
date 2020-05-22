// Types
import { RemoteDataState, LocalStorage } from 'src/types'

export const normalizeResources = (state: LocalStorage) => {
  return {
    variables: state.resources.variables,
    orgs: normalizeOrgs(state.resources.orgs),
    snippits: state.resources.snippits
  }
}

const normalizeOrgs = (orgs: LocalStorage['resources']['orgs']) => {
  return {
    ...orgs,
    org: orgs.org,
    status: RemoteDataState.NotStarted,
  }
}
