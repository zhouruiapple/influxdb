// Libraries
import React, {FC, useContext, useCallback} from 'react'

// Components
import {InfluxColors, List, Gradients} from '@influxdata/clockface'
import {PipeContext} from 'src/notebooks/context/pipe'

// Constants
import {WRITE_INTERVALS} from './index'

const ScheduleSelector: FC = () => {
  const {data, update} = useContext(PipeContext)
  const selectedSchedule = data.every

  const updateSchedule = useCallback(
    (every: string): void => {
      update({every})
    },
    [update]
  )

  return (
    <div className="write-bucket--block">
      <div className="write-bucket--block-title">Schedule</div>
      <List
        className="write-bucket--list"
        backgroundColor={InfluxColors.Obsidian}
      >
        {WRITE_INTERVALS.map(interval => (
          <List.Item
            key={interval.label}
            value={interval.every}
            onClick={updateSchedule}
            selected={interval.every === selectedSchedule}
            title={interval.every}
            gradient={Gradients.GundamPilot}
            wrapText={true}
          >
            <List.Indicator type="dot" />
            {`Every ${interval.every}`}
          </List.Item>
        ))}
      </List>
    </div>
  )
}

export default ScheduleSelector
