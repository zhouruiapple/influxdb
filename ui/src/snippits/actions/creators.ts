// Types
import {RemoteDataState, SnippitEntities} from 'src/types'
import {NormalizedSchema} from 'normalizr'

export const SET_SNIPPITS = 'SET_SNIPPITS'
export const ADD_SNIPPIT = 'ADD_SNIPPIT'
export const EDIT_SNIPPIT = 'EDIT_SNIPPIT'
export const REMOVE_SNIPPIT = 'REMOVE_SNIPPIT'

export type Action =
  | ReturnType<typeof setSnippits>
  | ReturnType<typeof addSnippit>
  | ReturnType<typeof editSnippit>
  | ReturnType<typeof removeSnippit>

export const setSnippits = (
  status: RemoteDataState,
  schema?: NormalizedSchema<SnippitEntities, string[]>
) =>
  ({
    type: SET_SNIPPITS,
    status,
    schema,
  } as const)

export const addSnippit = (schema: NormalizedSchema<SnippitEntities, string>) =>
  ({
    type: ADD_SNIPPIT,
    schema,
  } as const)

export const editSnippit = (schema: NormalizedSchema<SnippitEntities, string>) =>
  ({
    type: EDIT_SNIPPIT,
    schema,
  } as const)

export const removeSnippit = (id: string) =>
  ({
    type: REMOVE_SNIPPIT,
    id,
  } as const)
