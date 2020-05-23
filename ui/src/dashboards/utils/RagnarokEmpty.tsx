import React, {Component,PureComponent, RefObject, createRef} from 'react'


import {Page, Tabs, Table, Orientation, ComponentSize, IndexList, Alignment, SquareButton, IconFont, ConfirmationButton, ComponentColor, Appearance, Popover, PopoverPosition, Button, ButtonRef} from '@influxdata/clockface'

import {listServices,listInstances,listActivities} from 'src/dashboards/utils/ragnarok'
import FunctionCategory from 'src/timeMachine/components/fluxFunctionsToolbar/FunctionCategory'
import BucketAddDataButton from 'src/buckets/components/BucketAddDataButton'



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
    /*this.timerID = setInterval(
      () => this.tick(),
      1000
    );*/
  }

  componentWillUnmount() {
    //clearInterval(this.timerID);
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
      return <InstancesTable instances={this.state.instances} services={this.state.services}/>
    } else {
      return  <ActivitiesTable activities={this.state.activities}/>
    }
  }

}

function ServicesTable (props) {
  return <IndexList>
    <IndexList.Header>
      <IndexList.HeaderCell width="20%" columnName="Name"/>
      <IndexList.HeaderCell width="35%" columnName="Description"/>
      <IndexList.HeaderCell width="20%" columnName="Category"/>
      <IndexList.HeaderCell width="20%" columnName="Description"/>
      <IndexList.HeaderCell width="5%" columnName=""/>
    </IndexList.Header>
    <IndexList.Body columnCount={5} emptyState={null}>
     
        {props.services && props.services.map(s => (
           <IndexList.Row brighten={true} key={s.id}>
            <IndexList.Cell>{s.name}</IndexList.Cell>
            <IndexList.Cell>{s.description}</IndexList.Cell>
            <IndexList.Cell>{category(s.tags)}</IndexList.Cell>
            <IndexList.Cell>{formatTagsExcCat(s.tags)}</IndexList.Cell>
            <IndexList.Cell>{formatTagsExcCat(s.tags)}</IndexList.Cell>
            <IndexList.Cell alignment={Alignment.Right} revealOnHover={true}>
              <Button icon={IconFont.Plus} size={ComponentSize.ExtraSmall} text="Instance" color={ComponentColor.Secondary}/>
            </IndexList.Cell>
          </IndexList.Row>
      ))}
    </IndexList.Body>
    </IndexList>
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

function getService (instance, services) : any {
  for (const s of services) {
    if (s.id == instance.serviceId) {
      return s
    }
  }
  return null
}

function InstancesTable (props) {
  return <IndexList>
    <IndexList.Header>
      <IndexList.HeaderCell width="45%" columnName="Name"/>
      <IndexList.HeaderCell width="25%" columnName="Type"/>
      <IndexList.HeaderCell width="20%" columnName="Status"/>
      <IndexList.HeaderCell width="5%" columnName=""/>
      <IndexList.HeaderCell width="5%" columnName=""/>
    </IndexList.Header>
    <IndexList.Body columnCount={5} emptyState={null}> 
      {props.instances && props.instances.map(i => (
           <IndexList.Row brighten={true} key={i.id}>
            <IndexList.Cell>{i.name}</IndexList.Cell>
            <IndexList.Cell>{i.serviceType}</IndexList.Cell>
            <IndexList.Cell>{i.status}</IndexList.Cell>
            <IndexList.Cell alignment={Alignment.Center} revealOnHover={true}>
              <ActionsButton
                onAddCollector={()=>{}}
                onAddLineProtocol={()=>{}}
                onAddClientLibrary={()=>{}}
                onAddScraper={()=>{}}
                service={getService(i,props.services)}
            />
            </IndexList.Cell>
            <IndexList.Cell alignment={Alignment.Right} revealOnHover={true}>
                <ConfirmationButton
                icon={IconFont.Trash}
                testID="delete-token"
                size={ComponentSize.ExtraSmall}
                text="Delete"
                confirmationLabel="Really delete this instance?"
                confirmationButtonText="Confirm"
                confirmationButtonColor={ComponentColor.Danger}
                popoverAppearance={Appearance.Outline}
                popoverColor={ComponentColor.Danger}
                color={ComponentColor.Danger}
                onConfirm={()=>{}}
              />
            </IndexList.Cell>
          </IndexList.Row>
      ))}
    </IndexList.Body>
    </IndexList>
}

function ActivitiesTable (props) {

  return <IndexList>
    <IndexList.Header>
      <IndexList.HeaderCell width="20%" columnName="Time"/>
      <IndexList.HeaderCell width="20%" columnName="Instance"/>
      <IndexList.HeaderCell width="20%" columnName="Type"/>
      <IndexList.HeaderCell width="15%" columnName="Action"/>
      <IndexList.HeaderCell width="10%" columnName="Behaviour"/>
      <IndexList.HeaderCell width="10%" columnName="Status"/>
      <IndexList.HeaderCell width="5%" columnName=""/>
    </IndexList.Header>
    <IndexList.Body columnCount={7} emptyState={null}> 
    {props.activities && props.activities.map(a => (
           <IndexList.Row brighten={true} key={a.activityId}>
            <IndexList.Cell>{(new Date(a.timestamp)).toUTCString()}</IndexList.Cell>
            <IndexList.Cell>{a.instanceName}</IndexList.Cell>
            <IndexList.Cell>{a.serviceName}</IndexList.Cell>
            <IndexList.Cell>{a.operationName}</IndexList.Cell>
            <IndexList.Cell>{a.activityType}</IndexList.Cell>
            <IndexList.Cell>{a.status}</IndexList.Cell>
            <IndexList.Cell alignment={Alignment.Right} revealOnHover={true}>
                <ConfirmationButton
                icon={IconFont.Stop}
                testID="delete-token"
                size={ComponentSize.ExtraSmall}
                text="Stop"
                confirmationLabel="Really stop this activity?"
                confirmationButtonText="Confirm"
                confirmationButtonColor={ComponentColor.Danger}
                popoverAppearance={Appearance.Outline}
                popoverColor={ComponentColor.Danger}
                color={ComponentColor.Danger}
                onConfirm={()=>{}}
              />
            </IndexList.Cell>
          </IndexList.Row>
      ))}
    </IndexList.Body>
    </IndexList>
}


interface Props {
  onAddCollector: () => void
  onAddLineProtocol: () => void
  onAddClientLibrary: () => void
  onAddScraper: () => void
  service: any
}

class ActionsButton extends PureComponent<Props> {
  private triggerRef: RefObject<ButtonRef> = createRef()

  public render() {
    const {
      onAddCollector,
    } = this.props

    return (
      <>
        <Popover
          color={ComponentColor.Secondary}
          appearance={Appearance.Outline}
          position={PopoverPosition.ToTheRight}
          triggerRef={this.triggerRef}
          distanceFromTrigger={8}
          contents={onHide => (
            <div className="bucket-add-data" onClick={onHide}>
            {this.props.service.actions.map(a=>(
              <div key={a.name} className="bucket-add-data--option" onClick={onAddCollector}>
                <div className="bucket-add-data--option-header">
                  {a.name}
                </div>
                <div className="bucket-add-data--option-desc">
                  {a.description}
                </div>
              </div>
            ))}
            </div>
          )}
        />
        <Button
          ref={this.triggerRef}
          text="Action"
          testID="add-data--button"
          icon={IconFont.Zap}
          size={ComponentSize.ExtraSmall}
          color={ComponentColor.Secondary}
        />
      </>
    )
  }
}



