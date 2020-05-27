import React from 'react'
import {Dropdown, Popover, PopoverPosition, PopoverInteraction} from '@influxdata/clockface'
import { createGzip } from 'zlib'
import { Proposed } from 'monaco-languageclient/lib/services'
import { prependListener } from 'cluster'
import THEME_NAME from 'src/external/monaco.flux.theme'

export type Props = {
  services: any,
  instances: any,
  onClick: (event:any) => void
}


type NameIDPair = {
  name: string,
  id: string,
  service: any,
  instance: any,
  action: any,
}
type CategorizedList = {
  name: string,
  list: NameIDPair[],
 // instancesList: any
}

type CategoryOrService = {
    category: boolean,
    name: string,
    id: string,
    service: any,
    instance: any,
    action: any
}

export const RagnarokServicesDropdown = ({services,instances,onClick}: Props) => {

  // categorize them here
  var categorizedServices: CategorizedList[] = []

  if (services != null) {
        for (const s of services) {
            if (s.tags != null) {
                for (const t of s.tags) {
                    if (t.startsWith("Category=")) {
                        var cat = t.substring(9)


                        for (const action of s.actions) {
                            if (action.enablingStatuses == null || 
                                action.enablingStatuses.length == 0 || 
                                action.enablingStatuses.includes(s.initialStatus)) {

                                let added = false
                                for (const cs of categorizedServices) {
                                    if (cs.name == cat) {
                                        cs.list.push({name:action.name + " new " + s.name,id:s.id,service:s,action:action,instance:null})
                                        added = true
                                        //console.log("added",s.name,"to",cs.name)
                                        break
                                    } else {
                                        //console.log("didn't add",s.name,"to",cs.name)
                                    }
                                }
                                if (!added) {
                                    //console.log("new category",cat)
                                    categorizedServices.push({name:cat,list:[{name:action.name + " new " + s.name,id:s.id,service:s,action:action,instance:null}]})
                                }

                           } else {
                               //console.log("not adding",s.name,action.name,action.enablingStatuses,action.enablingStatuses.length,action.enablingStatuses[0],action.enablingStatuses[0] == "Ready",s.initialStatus)
                           }
                        }

                        if (instances != null) {
                            for (const i of instances) {
                                if (i.serviceId == s.id) {
                                    for (const a of s.actions) {
                                        if (a.enablingStatuses == null || 
                                            a.enablingStatuses.length == 0 || 
                                            a.enablingStatuses.includes(s.initialStatus)) {
                                                // skip these 'static' initializer actions
                                            } else if (a.enablingStatuses != null && a.enablingStatuses.includes(i.status)) {
                                            // need to add this service,instance,action tuple
                                            
                                            let added = false
                                            for (const cs of categorizedServices) {
                                                if (cs.name == cat) {
                                                    cs.list.push({name:a.name + " on '" + i.name + "' (" + s.name+")",id:s.id,service:s,action:a,instance:i})
                                                    added = true
                                                    //console.log("added",s.name,"to",cs.name)
                                                    break
                                                } else {
                                                    //console.log("didn't add",s.name,"to",cs.name)
                                                }
                                            }
                                            if (!added) {
                                                //console.log("new category",cat)
                                                categorizedServices.push({name:cat,list:[{name:a.name + " on '" + i.name + "' (" + s.name+")",id:s.id,service:s,action:a,instance:i}]})
                                            }
                                        }
                                    }
                                }
                            }
                        }
                    }
                }
            }
        }
    }



   // I NEED TO SORT THEM
   // I NEED TO REMOVE ACTIONS THAT INITIALIZER ACTIONS
    
    
  if (instances != null) {
    for (const i of instances) {
    }    
  }


  var toRender: CategoryOrService[] = []
  categorizedServices.map(cs=>{
      toRender.push({
        category: true,
        name: cs.name,
        id: cs.name,
        service:null,
        action:null,
        instance:null,
      })
      cs.list.map(m => {
          toRender.push({
              category:false,
              name: m.name,
              id: m.id,
              service:m.service,
              action:m.action,
              instance:m.instance,
          })
      })
  })

  //console.log("cat services are",categorizedServices)
  //console.log("render service",toRender)

  return <Dropdown
    button={(active, onClick) => (
        <Dropdown.Button
        active={active}
        onClick={onClick}
        >Services</Dropdown.Button>
    )}
    menu={(onCollapse) => (
        <Dropdown.Menu
        className="ragnarok-services"
        noScrollX={false}
        noScrollY={false}
        onCollapse={onCollapse}
        >
            {toRender.map(r => (
              <RenderCategory key={r.name} text={r.name} name={r.name} category={r.category} id={r.id} service={r.service} action={r.action} instance={r.instance} onClick={onClick}/>
            ))}
        </Dropdown.Menu>)}
  />
}

function RenderCategory (props) {
    if (props.category) {
        return <Dropdown.Divider key={props.name} text={props.name}/>
    }
    else {
        return <Dropdown.Item
            key={props.id}
            value={props.name}
            onClick={() => {
                props.onClick({id: props.id, name: props.name, service:props.service, action:props.action, instance:props.instance})
              }}
            >
                {props.name}
            </Dropdown.Item>
    }
}