import React, {Component,PureComponent} from 'react'


import {Page, Tabs, Table, Orientation, ComponentSize, BorderType} from '@influxdata/clockface'

import {listServices,listInstances,listActivities} from 'src/dashboards/utils/ragnarok'
import FunctionCategory from 'src/timeMachine/components/fluxFunctionsToolbar/FunctionCategory'



export default class RangarokEmpty extends Component {

  state = {
    activeTab : "types",
    services: null,
    instances: null,
    activities: null,
  }

  timerID: any

  componentDidMount() {
    this.tick()
    this.timerID = setInterval(
      () => this.tick(),
      1000
    );
  }

  componentWillUnmount() {
    clearInterval(this.timerID);
  }

  tick() {
    listServices().then((services)=>{
      this.setState({services:services})
    })
    listInstances().then((instances)=>{
      this.setState({instances:instances})
    })
    listActivities().then((activities)=>{
      this.setState({activities:activities})
    })
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
            <Tabs.Container
             orientation={Orientation.Horizontal}>

          <Tabs
            orientation={Orientation.Horizontal}
            size={ComponentSize.Large}
            dropdownBreakpoint={872}
          >
            
            <Tabs.Tab
              key="types"
              text="Types"
              id="types"
              active={this.state.activeTab == "types"}
              onClick={this.typesActive}
            />    

            <Tabs.Tab
              key="instances"
              text="Instances"
              id="instances"
              active={this.state.activeTab == "instances"}
              onClick={this.instancesActive}
            />    

            <Tabs.Tab
              key="activites"
              text="Activities"
              id="activities"
              active={this.state.activeTab == "activities"}
              onClick={this.activitisActive}
            />   
            </Tabs> 
             <Tabs.TabContents>{this.tabContents}</Tabs.TabContents>
          
          </Tabs.Container>

          </Page.Contents>
        </Page>
    )    
  }

  private typesActive = () => {
    this.setState({activeTab:"types"})
  }

  private instancesActive = () => {
    this.setState({activeTab:"instances"})
  }

  private activitisActive = () => {
    this.setState({activeTab:"activities"})
  }

  private get tabContents(): JSX.Element {
    const {activeTab} = this.state

    if (activeTab === "types") {
      return <ServicesTable services={this.state.services}/>
    } else if (activeTab === "instances") {
      return <InstancesTable instances={this.state.instances}/>
    } else {
      return  <ActivitiesTable activities={this.state.activities}/>
    }
  }

}

function ServicesTable (props) {
  return <Table><Table.Header>
    <Table.Row>
      <Table.HeaderCell>Name</Table.HeaderCell>
      <Table.HeaderCell>Description</Table.HeaderCell>
      <Table.HeaderCell>Category</Table.HeaderCell>
      <Table.HeaderCell>Tags</Table.HeaderCell>
    </Table.Row>
  </Table.Header>
  <Table.Body>
    {props.services && props.services.map(s => (
      <Table.Row key={s.id}>
        <Table.Cell>{s.name}</Table.Cell>
        <Table.Cell>{s.description}</Table.Cell>
        <Table.Cell>{category(s.tags)}</Table.Cell>
        <Table.Cell>{formatTagsExcCat(s.tags)}</Table.Cell>
      </Table.Row>
    ))}
  </Table.Body>
  </Table>
}

function formatTagsExcCat (tags) {
  let result = ""
  tags.forEach(t => {
    if (!t.startsWith("Category=")) {
      if (result.length > 0) {
        result += ", "
      }
      result += t
    }
  });
  return result
}

function category (tags) {
  let result = ""
  tags.forEach(t => {
    if (t.startsWith("Category=")) {
      result =  t.substring(9)
    }
  });
  return result
}

function InstancesTable (props) {
  return <Table><Table.Header>
    <Table.Row>
      <Table.HeaderCell>Name</Table.HeaderCell>
      <Table.HeaderCell>Type</Table.HeaderCell>
      <Table.HeaderCell>Status</Table.HeaderCell>
    </Table.Row>
  </Table.Header>
  <Table.Body>
    {props.instances && props.instances.map(i => (
      <Table.Row key={i.id}>
        <Table.Cell>{i.name}</Table.Cell>
        <Table.Cell>{i.serviceType}</Table.Cell>
        <Table.Cell>{i.status}</Table.Cell>
      </Table.Row>
    ))}
  </Table.Body>
  </Table>
}

function ActivitiesTable (props) {
  return <Table><Table.Header>
    <Table.Row>
      <Table.HeaderCell>Time</Table.HeaderCell>
      <Table.HeaderCell>Instance</Table.HeaderCell>
      <Table.HeaderCell>Type</Table.HeaderCell>
      <Table.HeaderCell>Action</Table.HeaderCell>
      <Table.HeaderCell>Behaviour</Table.HeaderCell>
      <Table.HeaderCell>Status</Table.HeaderCell>
    </Table.Row>
  </Table.Header>
  <Table.Body>
    {props.activities && props.activities.map(a => (
      <Table.Row key={a.activityId}>
        <Table.Cell>{(new Date(a.timestamp)).toUTCString()}</Table.Cell>
        <Table.Cell>{a.instanceName}</Table.Cell>
        <Table.Cell>{a.serviceName}</Table.Cell>
        <Table.Cell>{a.operationName}</Table.Cell>
        <Table.Cell>{a.activityType}</Table.Cell>
        <Table.Cell>{a.status}</Table.Cell>
      </Table.Row>
    ))}
  </Table.Body>
  </Table>
}

