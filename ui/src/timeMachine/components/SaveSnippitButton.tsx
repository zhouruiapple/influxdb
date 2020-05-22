import React, {FC} from 'react'
import {connect} from 'react-redux'
import uuid from 'uuid'

// Components
import {Button, IconFont} from '@influxdata/clockface'
import {createSnippit} from 'src/snippits/actions/thunks'

// Utils
import {getActiveQuery} from 'src/timeMachine/selectors'

// Types
import {AppState} from 'src/types'

interface StateProps {
  activeQueryText: string
}

interface DispatchProps {
  onSave: typeof createSnippit
}

const SaveSnippitButton: FC<StateProps & DispatchProps> = ({onSave, activeQueryText}) => <Button
  titleText="Save code as snippit"
  text="Save Snippit"
  icon={IconFont.Export}
  onClick={() => onSave({id: uuid.v4(), name: `${activeQueryText.slice(0, 30)}...`, code: activeQueryText})} />

const mstp = (state: AppState) => {
  const activeQueryText = getActiveQuery(state).text

  return {activeQueryText}
}

const mdtp = {
  onSave: createSnippit
}

export default connect<{}, DispatchProps, {}>(
  mstp,
  mdtp
)(SaveSnippitButton)
