import React, {FC, useContext} from 'react'
import {PipeProp} from 'src/notebooks'
import {NotebookContext} from 'src/notebooks/context/notebook'

import {NotebookPanel} from 'src/notebooks/components/Notebook'

const ExampleView: FC<PipeProp> = ({index, remove, moveUp, moveDown}) => {
  const {pipes, updatePipe} = useContext(NotebookContext)
  const pipe = pipes[index]

  const handleTitleChange = (title: string): void => {
    const updatedPipe = {...pipe, title}
    updatePipe(index, updatedPipe)
  }

  return (
    <NotebookPanel
      id={`pipe${index}`}
      onMoveUp={moveUp}
      onMoveDown={moveDown}
      onRemove={remove}
      title={pipe.title}
      onTitleChange={handleTitleChange}
    >
      <h1>{pipe.text}</h1>
    </NotebookPanel>
  )
}

export default ExampleView
