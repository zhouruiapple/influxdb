// Libraries
import React, {PureComponent, MouseEvent} from 'react'
import {connect} from 'react-redux'

import {Tabs, Orientation, ComponentSize, TextBlock, InfluxColors} from '@influxdata/clockface'

import { uniqueNamesGenerator, Config, adjectives, colors, animals } from 'unique-names-generator';

import {
    ComponentColor,
    Input,
    Button,
    Overlay,
    Form,  
  } from '@influxdata/clockface'

interface Props {
    isVisible: boolean,
    service: any,
    action: any,
    instance: any,
    instanceName: any,
    onCancel: () => void
    onApply: (obj: any) => void
}

export class RagnarokActionParametersForm extends PureComponent<Props> {

    state = {
        destinationBucket: this.props.action.output ? this.props.action.output.defaultDatabase : '',
        destinationMeasurement: this.props.action.output ? this.props.action.output.defaultMeasurement : '',
        outputTags: this.props.action.output ? this.props.action.output.defaultTags : '',
        repeat: '0s',
        parameterValues: this.getInitialParameterValues(),
        general: true,
        instanceName: generateInstanceName(),
        instance: this.props.instance,
    }

    public render() {

        return (
        <Overlay visible={this.props.isVisible}>
            <Overlay.Container maxWidth={600}>
            <Overlay.Header title={this.props.action.name + " with " + this.props.service.name}/>
            <Overlay.Body key="rapf">
                <TextBlock key="rapf-1" text={`Service: ${this.props.service.description}`}  />
                {this.props.instance &&
                    <TextBlock key="rapf-2" text={`Instance: ${this.props.instance.name}`} textColor={InfluxColors.Sidewalk}/>
                }
                <TextBlock key="rapf-3" text={`Action: ${this.props.action.description}`} />
                <Tabs.Container orientation={Orientation.Horizontal}>
                    <Tabs orientation={Orientation.Horizontal}size={ComponentSize.ExtraSmall} dropdownBreakpoint={872}>
                        <Tabs.Tab
                        key="General"
                        text="General"
                        id="General"
                        active={this.state.general}
                        onClick={()=>{this.setState({general:true})}}
                        />    

                        <Tabs.Tab
                        key="Advanced"
                        text="Advanced"
                        id="Advanced"
                        active={!this.state.general}
                        onClick={()=>{this.setState({general:false})}}
                        />    
                    </Tabs> 
                    <Tabs.TabContents>

                        {this.state.general && this.state.instance == null && [
                            <Form.Label key="instance-name" required={true} label="Instance name"/>,
                            <Input key="instance-name-f" value={this.state.instanceName} placeholder="instance name" onChange={this.updateInstanceName}/>,
                            <Form.HelpText key="instance-name-h" text="Name for this new service instance"></Form.HelpText>
                        ]}

                        {this.state.general && this.props.action.parameters.map(p => (
                            [<Form.Label key={p.name+"-l"} required={true} label={p.name}/>,
                                <Input key={p.name+"-f"} value={this.state.parameterValues[p.name]} placeholder={p.name} onChange={(e)=>{this.updateParam(p.name,e)}} />,
                                <Form.HelpText key={p.name+"-h"} text={p.description}></Form.HelpText>]
                            ))}
                        {this.state.general && this.props.action.output && [
                            <Form.Divider key="rapf-4" lineColor={ComponentColor.Primary}/>,
                            <Form.Label key="rapf-5" required={true} label="Destination Bucket"/>,
                            <Input key="rapf-6" value={this.state.destinationBucket} placeholder="Destination Bucket" onChange={this.updateDestinationBucket} />,
                            <Form.HelpText key="rapf-7" text="The bucket to write resuls to"/>,
                            <Form.Divider key="rapf-8" lineColor={ComponentColor.Primary}/>,
                            <Form.Label key="rap-9" required={true} label="Destination Measurement"/>,
                            <Input key="rapf-10" value={this.state.destinationMeasurement} placeholder="Destination Measurement" onChange={this.updateDestinationMeasurement} />,
                            <Form.HelpText key="rapf-11" text="The measurement to write resuls to"/>
                        ]}
                        {!this.state.general &&  this.props.action.output && [
                            <Form.Divider key="rapf-12" lineColor={ComponentColor.Primary}/>,
                            <Form.Label key="rapf-13" required={false} label="Output Tags"/>,
                            <Input key="rapf-14" value={this.state.outputTags} placeholder="Output Tags" onChange={this.updateOutputTags}/>,
                            <Form.HelpText key="rapf-15" text="Any additional tags to attach to the results"/>
                        ]}
                        {!this.state.general && [
                            <Form.Divider key="rapf-16" lineColor={ComponentColor.Primary}/>,
                            <Form.Label key="rapf-17" required={false} label="Repeat Every"/>,
                            <Input key="rapf-18" value={this.state.repeat} placeholder="How often to repeat" onChange={this.updateRepeat}/>,
                            <Form.HelpText key="rapf-19" text="If required, how often should this action to be repeated?"/>
                        ]}
                    </Tabs.TabContents>
                </Tabs.Container>
            </Overlay.Body>
            <Overlay.Footer>
                <Button text="Apply" onClick={()=>this.props.onApply(this.state)} />
                <Button text="Cancel" onClick={this.props.onCancel} />
            </Overlay.Footer>
            </Overlay.Container>
        </Overlay>
        )
    }


private getInitialParameterValues () {
    let result = {}
    this.props.action.parameters.forEach(p =>{
        result[p.name] = p.default ? p.default : ''
    })
    return result
}

private updateParam = (name, event) => {
    let p = {}
    Object.assign(p, this.state.parameterValues );
    p[name] = event.target.value
    this.setState({parameterValues:p}) 
  }

  private updateRepeat = (event) => {
    this.setState({repeat:event.target.value})
  }

  private updateInstanceName = (event) => {
    this.setState({instanceName:event.target.value})
  }

  private updateOutputTags = (event) => {
    this.setState({outputTags:event.target.value})
  }

  private updateDestinationBucket = (event) => {
    this.setState({ destinationBucket: event.target.value })
  }

  private updateDestinationMeasurement = (event) => {
    this.setState({ destinationMeasurement: event.target.value })
  }
}

function generateInstanceName () {
    const instanceName: string = uniqueNamesGenerator({
        dictionaries: [colors,adjectives, animals],
        separator: '-'
      });
  
      console.log('instance name',instanceName)
      return instanceName
}

  