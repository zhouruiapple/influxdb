// Libraries
import React, {FC, useState} from 'react'

// Components
import FluxFunctionsToolbar from 'src/timeMachine/components/fluxFunctionsToolbar/FluxFunctionsToolbar'
import VariableToolbar from 'src/timeMachine/components/variableToolbar/VariableToolbar'
import SnippitsToolbar from 'src/timeMachine/components/snippitsToolbar/SnippitsToolbar'
import FluxToolbarTab from 'src/timeMachine/components/FluxToolbarTab'

// Types
import {FluxToolbarFunction} from 'src/types'

interface Props {
  activeQueryBuilderTab: string
  onInsertFluxFunction: (func: FluxToolbarFunction) => void
  onInsertVariable: (variableName: string) => void
  onInsertSnippit: (snippitCode: string) => void
}

type FluxToolbarTabs = 'functions' | 'variables' | 'snippits'

const FluxToolbar: FC<Props> = ({
  activeQueryBuilderTab,
  onInsertFluxFunction,
  onInsertVariable,
  onInsertSnippit,
}) => {
  const [activeTab, setActiveTab] = useState<FluxToolbarTabs>('functions')

  const handleTabClick = (id: FluxToolbarTabs): void => {
    setActiveTab(id)
  }


  const activeToolbar = () => {
    switch (activeTab) {
      case 'variables':
        return <VariableToolbar onClickVariable={onInsertVariable} />
      case 'snippits':
        return <SnippitsToolbar onInsertSnippit={onInsertSnippit} />
      default:
        return <FluxFunctionsToolbar onInsertFluxFunction={onInsertFluxFunction} />

    }
  }

  return (
    <div className="flux-toolbar">
      <div className="flux-toolbar--tabs">
        <FluxToolbarTab
          id="functions"
          onClick={handleTabClick}
          name="Functions"
          active={activeTab === 'functions'}
          testID="functions-toolbar-tab"
        />
        <FluxToolbarTab
          id="snippits"
          onClick={handleTabClick}
          name="My Snippits"
          active={activeTab === 'snippits'}
          testID="functions-toolbar-tab"
        />
        {activeQueryBuilderTab !== 'customCheckQuery' && (
          <FluxToolbarTab
            id="variables"
            onClick={handleTabClick}
            name="Variables"
            active={activeTab === 'variables'}
          />
        )}
      </div>
      <div className="flux-toolbar--tab-contents">{activeToolbar()}</div>
    </div>
  )
}

export default FluxToolbar
