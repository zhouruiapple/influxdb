export const PIPE_DEFINITIONS = {}

export interface TypeRegistration {
  type: string // a unique string that identifies a pipe
  component: JSX.Element // the view component for rendering the interface
  button: string // a human readable string for appending the type
  empty: any // the default state for an add
}

export function register(definition: TypeRegistration) {
  if (PIPE_DEFINITIONS.hasOwnProperty(definition.type)) {
    throw new Exception(
      `Pipe of type [${definition.type}] has already been registered`
    )
  }

  PIPE_DEFINITIONS[definition.type] = {
    ...definition,
  }
}

// NOTE: this loads in all the modules under the local pipes directory
// to make it easier to add new types
const context = require.context('./pipes', true, /index\.(ts|tsx)$/)
context.keys().forEach(key => {
  context(key)
})
