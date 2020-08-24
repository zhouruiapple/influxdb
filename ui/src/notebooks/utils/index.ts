let CELL_GENERATOR_INDEX = 0

export const getHumanReadableName = (type: string): string => {
  ++CELL_GENERATOR_INDEX

  switch (type) {
    case 'data':
      return `Bucket ${CELL_GENERATOR_INDEX}`
    case 'writeData':
      return `Write to Bucket ${CELL_GENERATOR_INDEX}`
    case 'visualization':
      return `Visualization ${CELL_GENERATOR_INDEX}`
    case 'markdown':
      return `Markdown ${CELL_GENERATOR_INDEX}`
    case 'query':
      return `Flux Script ${CELL_GENERATOR_INDEX}`
    default:
      return `Cell ${CELL_GENERATOR_INDEX}`
  }
}
