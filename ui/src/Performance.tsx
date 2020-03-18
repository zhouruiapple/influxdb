import React from 'react'

import {performanceLineWriter} from 'src/utils/performanceLineWriter'

const INTERVAL = 120

export class PerformanceBoy extends React.Component {
  averageFPS: number
  cummulativeFPS: number[]
  frameStartTime: number
  lastFrameEndTime: number
  startTime: number
  timeInCurrentFrame: number

  constructor(props) {
    super(props)
    this.startTime = performance.now()

    this.averageFPS = 0
    this.currentFPS = 0
    this.cummulativeFPS = Array(INTERVAL)
    this.maxFPS = 0

    this.frameStartTime = performance.now()
    this.lastFrameEndTime = this.frameStartTime
    this.timeInCurrentFrame = this.frameStartTime

    this.maxTimeInFrame = 0
    this.totalFrames = 0
    this.totalTimeInFrame = 0

    this.appPerformance = {
      currentFPS: this.currentFPS,
      maxFPS: this.maxFPS,
      totalFrames: this.totalFrames,
      averageFPS: this.averageFPS,
      timeInCurrentFrame: this.timeInCurrentFrame,
      maxTimeInFrame: this.maxTimeInFrame,
    }

    this.state = {
      currentFPS: this.currentFPS,
    }
  }

  public componentDidMount() {
    document.addEventListener('visibilitychange', this.handleVisibilityChange)
    this.updateFPSCounter()
    this.writeToDatabase()
  }

  handleVisibilityChange = () => {
    // When the document becomes invisible, total frames stop being counted.
    // FPS is calculated as (total frames / time elapsed since start).
    // Since frames aren't counted, but time elapsed keeps increasing, we want to reset counting state.
    if (!document.hidden) {
      this.totalFrames = 0
      this.startTime = performance.now()
    }
  }

  updateFPSCounter = () => {
    this.frameStartTime = performance.now()
    this.totalFrames++

    this.currentFPS =
      (this.totalFrames / (this.frameStartTime - this.startTime)) * 1000

    if (this.currentFPS > this.maxFPS) {
      this.maxFPS = this.currentFPS
    }

    this.cummulativeFPS.push(this.currentFPS)

    this.timeInCurrentFrame = performance.now() - this.lastFrameEndTime
    this.lastFrameEndTime = performance.now()
    this.totalTimeInFrame += this.timeInCurrentFrame
    if (this.timeInCurrentFrame > this.maxTimeInFrame) {
      this.maxTimeInFrame = this.timeInCurrentFrame
    }

    this.setState({
      currentFPS: this.currentFPS,
    })
    requestAnimationFrame(this.updateFPSCounter)
  }

  writeToDatabase = () => {
    if (this.totalFrames % INTERVAL === 0) {
      // https://v2.docs.influxdata.com/v2.0/write-data/#influxdb-api
      this.averageFPS =
        this.cummulativeFPS.reduce(
          (total, currentFPS) => total + currentFPS,
          0
        ) / INTERVAL

      this.appPerformance = {
        currentFPS: this.currentFPS,
        maxFPS: this.maxFPS,
        totalFrames: this.totalFrames,
        averageFPS: this.averageFPS,
        timeInCurrentFrame: this.timeInCurrentFrame,
        maxTimeInFrame: this.maxTimeInFrame,
      }

      performanceLineWriter(this.appPerformance)

      this.maxFPS = this.currentFPS
      this.cummulativeFPS = Array(INTERVAL)
      this.maxTimeInFrame = this.timeInCurrentFrame
    }
    requestAnimationFrame(this.writeToDatabase)
  }

  public render() {
    return (
      <>
        <div style={{zIndex: 10000}}>
          fps: {this.currentFPS}
          <br />
          max: {this.maxFPS}
          <br />
          totalFrames: {this.totalFrames}
          <br />
          time in frame: {this.timeInCurrentFrame} ms
          <br />
          average time per frame:{' '}
          {(this.totalTimeInFrame / this.totalFrames)} ms
          <br />
          worst time in frame: {this.maxTimeInFrame} ms
        </div>
        {this.props.children}
      </>
    )
  }
}
