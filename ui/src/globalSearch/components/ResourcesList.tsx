import React, {FunctionComponent} from 'react'
import {EmptyState, IndexList, Alignment} from '@influxdata/clockface'
import {connect} from 'react-redux'
import uuid from 'uuid'

import {ComponentSize} from 'src/clockface'
import {Organization, AppState} from 'src/types'

interface StateProps {
  globalSearch: any[]
  org: Organization
}

interface OwnProps {
  searchTerm?: string
}

type Props = OwnProps & StateProps

const ResourcesList: FunctionComponent<Props> = props => {
  const {globalSearch} = props

  console.log('results', globalSearch)

  return (
    <IndexList>
      <IndexList.Body
        emptyState={
          <EmptyState size={ComponentSize.Small}>
            <EmptyState.Text text="No resources match your search term" />
          </EmptyState>
        }
        columnCount={3}
      >
        {globalSearch.map(resource => (
          <IndexList.Row key={uuid()}>
            <IndexList.Cell>
              <a href="#">{resource.Name}</a>
            </IndexList.Cell>
            <IndexList.Cell>{resource.Description}</IndexList.Cell>
            <IndexList.Cell alignment={Alignment.Right}>
              {resource.IndexType}
            </IndexList.Cell>
          </IndexList.Row>
        ))}
      </IndexList.Body>
    </IndexList>
  )
}

const mstp = ({globalSearch, orgs: {org}}: AppState): StateProps => ({
  globalSearch: globalSearch.list,
  org,
})

export default connect<StateProps, {}, {}>(
  mstp,
  null
)(ResourcesList)
