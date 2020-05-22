// Libraries
import {schema} from 'normalizr'

// Types
import {ResourceType} from 'src/types'

/* Authorizations */

// Defines the schema for the "snippit" resource
export const snippitSchema = new schema.Entity(ResourceType.Snippits)
export const arrayOfSnippits = [snippitSchema]
