// Libraries
import React, {FC, useState} from 'react'
import {connect} from 'react-redux'

// Components
import FluxToolbarSearch from 'src/timeMachine/components/FluxToolbarSearch'
import {DapperScrollbars, ComponentSize, EmptyState} from '@influxdata/clockface'
import SnippitItem from './SnippitItem'

// Types
import {AppState, Snippit} from 'src/types'


interface StateProps {
  snippits: Snippit[]
}

interface OwnProps {
  onInsertSnippit: (snippitCode: string) => void
}

type Props = StateProps & OwnProps


const SnippitsToolbar: FC<Props> = ({snippits, onInsertSnippit}) => {
  const [searchTerm, setSearchTerm] = useState('')
  console.log('snippits:', snippits)
  const filteredSnippits = snippits.filter(v => v.name.includes(searchTerm))

  let content: JSX.Element | JSX.Element[] = (
    <EmptyState size={ComponentSize.ExtraSmall}>
      <EmptyState.Text>No snippits match your search</EmptyState.Text>
    </EmptyState>
  )

  if (Boolean(filteredSnippits.length)) {
    content = filteredSnippits.map(s => <SnippitItem onClickSnippit={onInsertSnippit} snippit={s} key={s.id} />)
  }

  return (
    <>
      <FluxToolbarSearch
        onSearch={setSearchTerm}
        resourceName="Snippit"
      />
      <DapperScrollbars className="flux-toolbar--scroll-area">
        <div className="flux-toolbar--list">
          {content}
        </div>
      </DapperScrollbars>
    </>
  )
}

const mstp = (state: AppState) => {
  const byID = state.resources.snippits.byID
  const snippits = state.resources.snippits.allIDs.map(id => byID[id])

  return {snippits}
}

export default connect<StateProps>(
  mstp
)(SnippitsToolbar)
