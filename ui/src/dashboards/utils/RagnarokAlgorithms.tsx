import React from 'react'
import {Dropdown} from '@influxdata/clockface'

export interface RagnarokAlgorithm {
  id: string
  name: string
}

export type Props = {
  algorithms: RagnarokAlgorithm[],
  onClick: (any) => void
}

export const RagnarokAlgorithms = ({algorithms, onClick}: Props) => {
  return (
    <Dropdown
        button={(active, onClick) => (
          <Dropdown.Button
            active={active}
            onClick={onClick}
          >Algorithms</Dropdown.Button>
        )}
        menu={(onCollapse) => (
          <Dropdown.Menu
            className="ragnarok-algorithms"
            noScrollX={true}
            noScrollY={true}
            onCollapse={onCollapse}
          >
            <Dropdown.Divider text="Forecast" />
            {algorithms.map(algorithm => (
              <Dropdown.Item
                key={algorithm.id}
                value={algorithm.name}
                onClick={() => {
                  onClick({id: algorithm.id, name: algorithm.name})
                }}
              >
                {algorithm.name}
              </Dropdown.Item>
            ))}
          </Dropdown.Menu>)}
    />
  )
}
