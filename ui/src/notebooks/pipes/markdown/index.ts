import {register} from 'src/notebooks'
import MarkdownPanel from './MarkdownPanel'
import './style.scss'

let counter = 0

register({
  type: 'markdown',
  component: MarkdownPanel,
  button: 'Markdown',
  initial: () => ({
    title: `Markdown ${counter++}`,
    text: `Wooooooo`,
    mode: 'preview',
  }),
})
