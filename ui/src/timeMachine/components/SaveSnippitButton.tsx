import React, { FC } from 'react'
import { connect } from 'react-redux'

// Components
import { Button, IconFont } from '@influxdata/clockface'
import { createSnippit } from 'src/snippits/actions/thunks'

interface DispatchProps {
  onSave: typeof createSnippit
}

const SaveSnippitButton: FC<DispatchProps> = ({ onSave }) => <Button
  titleText="Save code as snippit"
  text="Save Snippit"
  icon={IconFont.Export}
  onClick={() => onSave({ name: "MySnippit", code: "im a script" })} />

const mdtp = {
  onSave: createSnippit
}

export default connect<{}, DispatchProps, {}>(
  null,
  mdtp
)(SaveSnippitButton)
