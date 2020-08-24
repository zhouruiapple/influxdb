import {register} from 'src/notebooks'
import View from './view'
import './style.scss'

register({
  type: 'writeData',
  family: 'outputs',
  priority: 1,
  component: View,
  button: 'Write to Bucket',
  initial: {
    bucketName: '',
    every: '1h',
    offset: '',
  },
})

export interface WriteInterval {
  label: string
  every: string
}

export const WRITE_INTERVALS: WriteInterval[] = [
  {
    label: '1m',
    every: '1m',
  },
  {
    label: '5m',
    every: '5m',
  },
  {
    label: '10m',
    every: '10m',
  },
  {
    label: '15m',
    every: '15m',
  },
  {
    label: '20m',
    every: '20m',
  },
  {
    label: '30m',
    every: '30m',
  },
  {
    label: '45m',
    every: '45m',
  },
  {
    label: '1h',
    every: '1h',
  },
  {
    label: '2h',
    every: '2h',
  },
  {
    label: '6h',
    every: '6h',
  },
  {
    label: '12h',
    every: '12h',
  },
  {
    label: '24h',
    every: '24h',
  },
  {
    label: '2d',
    every: '2d',
  },
  {
    label: '7d',
    every: '7d',
  },
  {
    label: '30d',
    every: '30d',
  },
]
