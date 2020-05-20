import React, {FC, useContext} from 'react'
import Pipe from 'src/notebooks/components/Pipe'
import {NotebookContext} from 'src/notebooks/context/notebook'

const PipeList: FC = () => {
  const {id, pipes, removePipe, movePipe} = useContext(NotebookContext)
  const _pipes = pipes.map((_, index) => {
    const remove = () => removePipe(index)
    const moveUp = () => movePipe(index, index - 1)
    const moveDown = () => movePipe(index, index + 1)

    const canBeMovedUp = index > 0
    const canBeMovedDown = index < pipes.length - 1
    const canBeRemoved = index !== 0

    return (
      <Pipe
        index={index}
        remove={canBeRemoved && remove}
        moveUp={canBeMovedUp && moveUp}
        moveDown={canBeMovedDown && moveDown}
        key={`pipe-${id}-${index}`}
      />
    )
  })

  return <>{_pipes}</>
}

export default PipeList
