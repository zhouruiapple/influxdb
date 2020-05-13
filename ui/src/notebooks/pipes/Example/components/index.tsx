import React, {FC, useContext} from 'react'
import {NotebookContext} from 'src/notebooks/context/notebook'
import NotebookPanel from 'src/notebooks/components/NotebookPanel'

interface Props {
    idx: number
}

const Visualization: FC<Props> = ({idx}) => {
  const {pipes, removePipe} = useContext(NotebookContext)

  const pipe = pipes[idx]
  const remove = idx ? () => removePipe(idx) : false

  return (
    <NotebookPanel
      onRemove={ remove }
      title={ pipe.title }
    >
        <h1>{ pipe.title }</h1>
    </NotebookPanel>
  )
}

export default Visualization
