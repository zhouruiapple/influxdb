// Libraries
import React, {FC, useContext, useState, ChangeEvent} from 'react'

// Types
import {PipeProp} from 'src/notebooks'

// Components
import {NotebookPanel} from 'src/notebooks/components/Notebook'
import MarkdownPanelModeToggle from './MarkdownPanelModeToggle'
import {MarkdownRenderer} from 'src/shared/components/views/MarkdownRenderer'

// Contexts
import {NotebookContext} from 'src/notebooks/context/notebook'

export type MarkdownPanelMode = 'edit' | 'preview'

const ExampleView: FC<PipeProp> = ({index, remove, moveUp, moveDown}) => {
  const {pipes, updatePipe} = useContext(NotebookContext)
  const pipe = pipes[index]

  const [mode, setMode] = useState<MarkdownPanelMode>(pipe.mode)
  // const [content, updateContent] = useState<string>(pipe.content)

  const controlsRight = (
    <MarkdownPanelModeToggle mode={mode} onToggleMode={setMode} />
  )

  const handleTextareaChange = (e: ChangeEvent<HTMLTextAreaElement>): void => {
    const updatedPipe = {...pipe, text: e.target.value}
    updatePipe(index, updatedPipe)
  }

  let body = (
    <div className="notebook-panel--markdown markdown-format">
      <MarkdownRenderer text={pipe.text} />
    </div>
  )

  if (mode === 'edit') {
    // const textAreaStyle = {height: `${previewHeight}px`}

    body = (
      <textarea
        className="notebook-panel--markdown-edit"
        value={pipe.text}
        onChange={handleTextareaChange}
        // style={textAreaStyle}
        autoFocus={true}
        autoComplete="off"
      />
    )
  }


  return (
    <NotebookPanel
      id={`pipe${index}`}
      onMoveUp={moveUp}
      onMoveDown={moveDown}
      onRemove={remove}
      controlsRight={controlsRight}
      title={pipe.title}
    >
      {body}
    </NotebookPanel>
  )
}

export default ExampleView
