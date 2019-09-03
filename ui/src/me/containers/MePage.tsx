// Libraries
import React, {PureComponent} from 'react'
import {connect} from 'react-redux'
import {get} from 'lodash'

// Components
import {
  Panel,
  FlexBox,
  FlexDirection,
  ComponentSize,
  AlignItems,
  Grid,
  Columns,
  InfluxColors,
} from '@influxdata/clockface'
import {Page} from 'src/pageLayout'
import Sidebar from 'src/me/components/Sidebar'
import Header from 'src/me/components/UserPageHeader'
import GettingStarted from 'src/me/components/GettingStarted'
import DashboardsList from 'src/me/components/DashboardsList'

// Types
import {AppState} from 'src/types'
import GetResources, {ResourceTypes} from 'src/shared/components/GetResources'

// Decorators
import {ErrorHandling} from 'src/shared/decorators/errors'

interface StateProps {
  me: AppState['me']
  orgName: string
}

@ErrorHandling
export class MePage extends PureComponent<StateProps> {
  public render() {
    const {me, orgName} = this.props

    return (
      <Page className="user-page" titleTag="My Account">
        <Header userName={me.name} orgName={orgName} />
        <Page.Contents fullWidth={false} scrollable={true}>
          <div className="col-xs-12">
            <Grid>
              <Grid.Row>
                <Grid.Column widthSM={Columns.Eight} widthMD={Columns.Nine}>
                  <FlexBox
                    direction={FlexDirection.Column}
                    margin={ComponentSize.Small}
                    alignItems={AlignItems.Stretch}
                    stretchToFitWidth={true}
                  >
                    <Panel>
                      <Panel.Header>
                        <Panel.Title>
                          Getting started with InfluxDB 2.0
                        </Panel.Title>
                      </Panel.Header>
                      <Panel.Body>
                        <GettingStarted />
                      </Panel.Body>
                    </Panel>
                    <Panel backgroundColor={InfluxColors.Kevlar}>
                      <Panel.Body>
                        <Grid>
                          <Grid.Row>
                            <Grid.Column widthSM={Columns.Four}>
                              <Panel>
                                <Panel.Header>
                                  <Panel.Title style={{fontSize: '16px'}}>
                                    Dashboards
                                  </Panel.Title>
                                </Panel.Header>
                                <Panel.Body>
                                  <GetResources
                                    resource={ResourceTypes.Dashboards}
                                  >
                                    <DashboardsList />
                                  </GetResources>
                                </Panel.Body>
                              </Panel>
                            </Grid.Column>
                            <Grid.Column widthSM={Columns.Four}>
                              <Panel>
                                <Panel.Header>
                                  <Panel.Title style={{fontSize: '16px'}}>
                                    Dashboards
                                  </Panel.Title>
                                </Panel.Header>
                                <Panel.Body>
                                  <GetResources
                                    resource={ResourceTypes.Dashboards}
                                  >
                                    <DashboardsList />
                                  </GetResources>
                                </Panel.Body>
                              </Panel>
                            </Grid.Column>
                            <Grid.Column widthSM={Columns.Four}>
                              <Panel>
                                <Panel.Header>
                                  <Panel.Title style={{fontSize: '16px'}}>
                                    Dashboards
                                  </Panel.Title>
                                </Panel.Header>
                                <Panel.Body>
                                  <GetResources
                                    resource={ResourceTypes.Dashboards}
                                  >
                                    <DashboardsList />
                                  </GetResources>
                                </Panel.Body>
                              </Panel>
                            </Grid.Column>
                          </Grid.Row>
                        </Grid>
                      </Panel.Body>
                    </Panel>
                  </FlexBox>
                </Grid.Column>
                <Grid.Column widthSM={Columns.Four} widthMD={Columns.Three}>
                  <Sidebar me={me} />
                </Grid.Column>
              </Grid.Row>
            </Grid>
          </div>
        </Page.Contents>
      </Page>
    )
  }
}

const mstp = (state: AppState): StateProps => {
  const {
    me,
    orgs: {org},
  } = state

  return {me, orgName: get(org, 'name', '')}
}

export default connect<StateProps>(
  mstp,
  null
)(MePage)
