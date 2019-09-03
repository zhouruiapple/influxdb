// Libraries
import React, {PureComponent} from 'react'

// Components
import Support from 'src/me/components/Support'
import LogoutButton from 'src/me/components/LogoutButton'
import {
  Panel,
  FlexBox,
  FlexDirection,
  ComponentSize,
  AlignItems,
} from '@influxdata/clockface'
import VersionInfo from 'src/shared/components/VersionInfo'
import Docs from 'src/me/components/Docs'

// Types
import {AppState} from 'src/types'

interface Props {
  me: AppState['me']
}

class ResourceLists extends PureComponent<Props> {
  public render() {
    return (
      <FlexBox
        direction={FlexDirection.Column}
        alignItems={AlignItems.Stretch}
        stretchToFitWidth={true}
        margin={ComponentSize.Small}
      >
        <Panel>
          <Panel.Header>
            <Panel.Title>My Account</Panel.Title>
            <LogoutButton />
          </Panel.Header>
        </Panel>
        <Docs />
        <Panel>
          <Panel.Header>
            <Panel.Title>Useful Links</Panel.Title>
          </Panel.Header>
          <Panel.Body>
            <Support />
          </Panel.Body>
          <Panel.Footer>
            <VersionInfo />
          </Panel.Footer>
        </Panel>
      </FlexBox>
    )
  }
}

export default ResourceLists
