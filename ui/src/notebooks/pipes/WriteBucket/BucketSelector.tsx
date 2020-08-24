// Libraries
import React, {FC, useEffect, useContext, useCallback} from 'react'

// Components
import {
  TechnoSpinner,
  ComponentSize,
  RemoteDataState,
  InfluxColors,
  List,
  Gradients,
} from '@influxdata/clockface'
import {BucketContext} from 'src/notebooks/context/buckets'
import {PipeContext} from 'src/notebooks/context/pipe'

// Types
import {Bucket} from 'src/types'

const BucketSelector: FC = () => {
  const {data, update} = useContext(PipeContext)
  const {buckets, loading} = useContext(BucketContext)

  const selectedBucketName = data.bucketName

  const updateBucket = useCallback(
    (updatedBucket: Bucket): void => {
      update({bucketName: updatedBucket.name})
    },
    [update]
  )

  useEffect(() => {
    // selectedBucketName will only evaluate false on the initial render
    // because there is no default value
    if (!!buckets.length && !selectedBucketName) {
      updateBucket(buckets[0])
    }
  }, [buckets, selectedBucketName, updateBucket])

  let body

  if (loading === RemoteDataState.Loading) {
    body = (
      <div className="write-bucket--list__empty">
        <TechnoSpinner strokeWidth={ComponentSize.Small} diameterPixels={32} />
      </div>
    )
  }

  if (loading === RemoteDataState.Error) {
    body = (
      <div className="write-bucket--list__empty">
        <p>Could not fetch Buckets</p>
      </div>
    )
  }

  if (loading === RemoteDataState.Done && selectedBucketName) {
    body = (
      <List
        className="write-bucket--list"
        backgroundColor={InfluxColors.Obsidian}
      >
        {buckets.map(bucket => (
          <List.Item
            key={bucket.name}
            value={bucket}
            onClick={updateBucket}
            selected={bucket.name === selectedBucketName}
            title={bucket.name}
            gradient={Gradients.GundamPilot}
            wrapText={true}
          >
            <List.Indicator type="dot" />
            {bucket.name}
          </List.Item>
        ))}
      </List>
    )
  }

  return (
    <div className="write-bucket--block">
      <div className="write-bucket--block-title">Destination</div>
      {body}
    </div>
  )
}

export default BucketSelector
