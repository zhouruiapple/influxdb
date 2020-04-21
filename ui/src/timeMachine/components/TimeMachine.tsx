// Libraries
import React, {FunctionComponent} from 'react'
import {connect} from 'react-redux'
import classnames from 'classnames'

// Components
import TimeMachineQueries from 'src/timeMachine/components/Queries'
import TimeMachineAlerting from 'src/timeMachine/components/TimeMachineAlerting'
import TimeMachineVis from 'src/timeMachine/components/Vis'
import TimeMachineRawVis from 'src/timeMachine/components/RawVis'
import ViewOptions from 'src/timeMachine/components/view_options/ViewOptions'
import TimeMachineCheckQuery from 'src/timeMachine/components/TimeMachineCheckQuery'

// Utils
import {getActiveTimeMachine} from 'src/timeMachine/selectors'

// Types
import {AppState, TimeMachineTab} from 'src/types'

interface StateProps {
  activeTab: TimeMachineTab
  isViewingVisOptions: boolean
}

const TimeMachine: FunctionComponent<StateProps> = ({
  activeTab,
  isViewingVisOptions,
}) => {
  const containerClassName = classnames('time-machine', {
    'time-machine--split': isViewingVisOptions,
  })

  let bottomContents: JSX.Element = null

  if (activeTab === 'alerting') {
    bottomContents = <TimeMachineAlerting />
  } else if (activeTab === 'queries') {
    bottomContents = <TimeMachineQueries />
  } else if (activeTab === 'customCheckQuery') {
    bottomContents = <TimeMachineCheckQuery />
  }

  return (
    <>
      {isViewingVisOptions && <ViewOptions />}
      <div className={containerClassName}>
        {bottomContents}
        <TimeMachineRawVis />
        <TimeMachineVis />
      </div>
    </>
  )
}

const mstp = (state: AppState) => {
  const {activeTab, isViewingVisOptions} = getActiveTimeMachine(state)

  return {activeTab, isViewingVisOptions}
}

export default connect<StateProps>(mstp)(TimeMachine)
