import {normalize} from 'normalizr'
import { Snippit, SnippitEntities} from 'src/types'
import { addSnippit } from './creators'
import {snippitSchema} from 'src/schemas/snippit'

export const createSnippit = (snippit: Snippit) => async (dispatch) => {
  const newSnippit = normalize<Snippit, SnippitEntities, string>(snippit, snippitSchema)

  dispatch(addSnippit(newSnippit))
}