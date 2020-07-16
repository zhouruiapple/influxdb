// Libraries
import React, {FC, useContext} from 'react'

// Components
import {Button, ComponentColor, ComponentStatus} from '@influxdata/clockface'

// Constants
import {NotebookContext} from 'src/notebooks/context/notebook.current'
import {ResultsContext} from 'src/notebooks/context/results'
import {PIPE_DEFINITIONS} from 'src/notebooks'

import {event} from 'src/cloud/utils/reporting'
import {isFlagEnabled} from 'src/shared/utils/featureFlag'

interface Props {
  index?: number
  onInsert?: () => void
  eventName: string
}

const AddButtons: FC<Props> = ({index, onInsert, eventName}) => {
  const {add, notebook} = useContext(NotebookContext)
  const results = useContext(ResultsContext)

  const dataSourceExistsInPipes = notebook.data.all.find(p => p.type === 'data')

  const pipes = Object.entries(PIPE_DEFINITIONS)
    .filter(
      ([_, def]) =>
        !def.disabled && (!def.featureFlag || isFlagEnabled(def.featureFlag))
    )
    .sort((a, b) => {
      const aPriority = a[1].priority || 0
      const bPriority = b[1].priority || 0

      if (aPriority === bPriority) {
        return a[1].button.localeCompare(b[1].button)
      }

      return bPriority - aPriority
    })
    .map(([type, def]) => {
      const buttonStatus =
        type === 'data' && dataSourceExistsInPipes
          ? ComponentStatus.Disabled
          : ComponentStatus.Default

      return (
        <Button
          key={def.type}
          text={def.button}
          status={buttonStatus}
          onClick={() => {
            let data = def.initial
            if (typeof data === 'function') {
              data = data()
            }

            onInsert && onInsert()

            event(eventName, {
              type: def.type,
            })

            const id = add(
              {
                ...data,
                type,
              },
              index
            )

            results.add(id)
          }}
          color={ComponentColor.Secondary}
        />
      )
    })

  return <>{pipes}</>
}

export default AddButtons
