import React, {Component,PureComponent} from 'react'
import {withRouter, WithRouterProps} from 'react-router'


import {Page} from '@influxdata/clockface'

import {Tabs, Orientation, ComponentSize} from '@influxdata/clockface'

interface OwnProps {
  activeTab: string
  orgID: string
}

type Props = OwnProps & WithRouterProps

class CatalogTabaNavigation extends PureComponent<Props> {
  public render() {
    const {activeTab, orgID, router} = this.props

    const handleTabClick = (id: string): void => {
      router.push(`/orgs/${orgID}/catalog/${id}`)
    }

    const tabs = [
      {
        text: 'Types',
        id: 'types',
      },
      {
        text: 'Instances',
        id: 'instances',
      },
      {
        text: 'Activities',
        id: 'activities',
      },
    ]

    const activeTabName = tabs.find(t => t.id === activeTab).text

    return (
      <Tabs
        orientation={Orientation.Horizontal}
        size={ComponentSize.Large}
        dropdownBreakpoint={872}
        dropdownLabel={activeTabName}
      >
        {tabs.map(t => {
          let tabElement = (
            <Tabs.Tab
              key={t.id}
              text={t.text}
              id={t.id}
              onClick={handleTabClick}
              active={t.id === activeTab}
            />
          )
          return tabElement
        })}
      </Tabs>
    )
  }
}

export default class RangarokEmpty extends Component {

  componentDidMount() {
    console.log("Empty Page mounted")
  }


  render() {

    return (
        <Page
          titleTag="Services"
        >
          <Page.Header fullWidth={false}>
            <Page.Title title="Services" />
          </Page.Header>

          <Page.Contents
            className="dashboards-index__page-contents"
            fullWidth={false}
            scrollable={true}
          >
            <CatalogTabaNavigation activeTab={"types"} orgID={"xxxxx"}/>
            <title>Hello</title>
          </Page.Contents>
        </Page>
    )    
  }
}
