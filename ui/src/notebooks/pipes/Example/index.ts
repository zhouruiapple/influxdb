import {register} from 'src/notebooks'
import View from './components/'
import './style.scss'

register({
  type: 'example',
  component: View,
  button: 'Append Example',
  empty: {
    text: 'Look at this example',
  },
})
