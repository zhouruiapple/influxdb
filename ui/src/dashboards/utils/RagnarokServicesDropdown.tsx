import React from 'react'
import {Dropdown, Popover, PopoverPosition, PopoverInteraction} from '@influxdata/clockface'
import { createGzip } from 'zlib'
import { Proposed } from 'monaco-languageclient/lib/services'

export type Props = {
  services: any,
  onClick: (event:any) => void
}


type NameIDPair = {
  name: string,
  id: string
}
type CategorizedList = {
  name: string,
  list: NameIDPair[]
}

type CategoryOrService = {
    category: boolean,
    name: string,
    id: string
}

export const RagnarokServicesDropdown = ({services,onClick}: Props) => {

  // categorize them here
  var categorizedServices: CategorizedList[] = []

  if (services != null) {
        for (const s of services) {
            if (s.tags != null) {
                for (const t of s.tags) {
                    if (t.startsWith("Category=")) {
                        var cat = t.substring(9)
                        let added = false
                        for (const cs of categorizedServices) {
                            if (cs.name == cat) {
                                cs.list.push({name:s.name,id:s.id})
                                added = true
                                //console.log("added",s.name,"to",cs.name)
                                break
                            } else {
                                //console.log("didn't add",s.name,"to",cs.name)
                            }
                        }
                        if (!added) {
                            //console.log("new category",cat)
                            categorizedServices.push({name:cat,list:[{name:s.name,id:s.id}]})
                        }
                    }
                }
            }
        }
    }

  var toRender: CategoryOrService[] = []
  categorizedServices.map(cs=>{
      toRender.push({
        category: true,
        name: cs.name,
        id: cs.name
      })
      cs.list.map(m => {
          toRender.push({
              category:false,
              name: m.name,
              id: m.id
          })
      })
  })

  //console.log("cat services are",categorizedServices)
  console.log("render service",toRender)

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
              <RenderCategory key={r.name} text={r.name} name={r.name} category={r.category} id={r.id} onClick={onClick}/>
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
                props.onClick({id: props.id, name: props.name})
              }}
            >
                {props.name}
            </Dropdown.Item>
    }
}