// Libraries
import React, {FC} from 'react'

// Components
import {SelectGroup} from '@influxdata/clockface'

// Types
import {MarkdownPanelMode} from './MarkdownPanel'

interface Props {
  mode: MarkdownPanelMode
  onToggleMode: (mode: MarkdownPanelMode) => void
}

const MarkdownPanelModeToggle: FC<Props> = ({mode, onToggleMode}) => {
  return (
    <SelectGroup>
      <SelectGroup.Option id="edit" active={mode === 'edit'} value="edit" onClick={onToggleMode}>Edit</SelectGroup.Option>
      <SelectGroup.Option id="active" active={mode === 'preview'} value="preview" onClick={onToggleMode}>Preview</SelectGroup.Option>
    </SelectGroup>
  )

}

export default MarkdownPanelModeToggle