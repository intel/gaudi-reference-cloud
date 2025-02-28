// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React, { useEffect, useRef, useState } from 'react'
import { BsChevronRight, BsChevronLeft } from 'react-icons/bs'
import type MediaFile from '../../store/models/MediaFile/MedialFile'
import './MediaCarousel.scss'
import { Button, Modal } from 'react-bootstrap'

interface MediaCarouselProps {
  mediaList: MediaFile[]
  className?: string
}

interface MediaCarouselCardProps {
  mediaFile: MediaFile | null
  index: number
  onClick: (media: MediaFile) => void
}

interface MediaCarouselModalProp {
  mediaFile: MediaFile | null
  showModal: boolean
  onClose: () => void
}

const MediaCarouselModal: React.FC<MediaCarouselModalProp> = ({ mediaFile, showModal, onClose }): JSX.Element => {
  if (mediaFile === null) {
    return <></>
  }

  let content = null

  if (mediaFile.type === 'img') {
    content = (
      <img
        className="mediaGalleryModal mediaGalleryImage"
        src={mediaFile.src}
        alt="Gallery image fullscreen"
        style={{
          backgroundImage: `url("${mediaFile.src}")`
        }}
      />
    )
  }

  return (
    <Modal show={showModal} onHide={onClose} fullscreen aria-label="Full screen media modal">
      <Modal.Header closeButton></Modal.Header>
      <Modal.Body className="d-flex align-items-center justify-content-center">{content}</Modal.Body>
    </Modal>
  )
}

const MediaCarouselCard: React.FC<MediaCarouselCardProps> = ({ mediaFile, index, onClick }): JSX.Element => {
  const videoRef = useRef<HTMLVideoElement>(null)

  if (mediaFile === null) {
    return <></>
  }

  if (mediaFile.type === 'img') {
    return (
      <img
        className="mediaGalleryCard mediaGalleryImage"
        onClick={() => {
          onClick(mediaFile)
        }}
        src={mediaFile.src}
        alt={`Gallery image at index ${index}`}
        style={{
          backgroundImage: `url("${mediaFile.src}")`
        }}
      />
    )
  }

  if (mediaFile.type === 'video') {
    return (
      <video
        id={`video-mini-${index}`}
        ref={videoRef}
        poster={mediaFile.poster}
        className="mediaGalleryCard"
        controls
        src={mediaFile.src}
        onClick={() => {
          const promise = videoRef?.current?.requestFullscreen()
          promise?.then(() => {}).catch(() => {})
        }}
      />
    )
  }

  return <></>
}

const MediaCarousel: React.FC<MediaCarouselProps> = ({ className = '', mediaList = [] }): JSX.Element => {
  const galleryRef = useRef<HTMLDivElement>(null)
  const buttonLeftRef = useRef<HTMLButtonElement>(null)
  const buttonRightRef = useRef<HTMLButtonElement>(null)

  const [selectedMedia, setSelectedMedia] = useState<MediaFile | null>(null)
  const [showModal, setShowModal] = useState(false)

  const onMediaClick = (media: MediaFile): void => {
    setSelectedMedia(media)
    setShowModal(true)
  }

  useEffect(() => {
    let interval: NodeJS.Timeout | string | number | undefined

    const leftMouseEnter = (): void => {
      interval = setInterval(() => {
        if (galleryRef.current?.scrollBy !== undefined) {
          galleryRef.current.scrollLeft -= 1
        }
      }, 5)
    }

    const leftMouseLeave = (): void => {
      clearInterval(interval)
    }

    const rightMouseEnter = (): void => {
      interval = setInterval(() => {
        if (galleryRef.current?.scrollLeft !== undefined) {
          galleryRef.current.scrollLeft += 1
        }
      }, 5)
    }

    const rightMouseLeave = (): void => {
      clearInterval(interval)
    }

    buttonLeftRef.current?.addEventListener('mouseenter', leftMouseEnter)
    buttonLeftRef.current?.addEventListener('mouseleave', leftMouseLeave)
    buttonRightRef.current?.addEventListener('mouseenter', rightMouseEnter)
    buttonRightRef.current?.addEventListener('mouseleave', rightMouseLeave)

    return () => {
      buttonLeftRef.current?.removeEventListener('mouseenter', leftMouseEnter)
      buttonLeftRef.current?.removeEventListener('mouseleave', leftMouseLeave)
      buttonRightRef.current?.removeEventListener('mouseenter', rightMouseEnter)
      buttonRightRef.current?.removeEventListener('mouseleave', rightMouseLeave)
    }
  }, [])

  const jumpToLeft = (): void => {
    if (galleryRef.current?.scrollLeft !== undefined) {
      galleryRef.current.scrollLeft -= 300
    }
  }

  const jumpToRight = (): void => {
    if (galleryRef.current?.scrollLeft !== undefined) {
      galleryRef.current.scrollLeft += 300
    }
  }

  return (
    <>
      <MediaCarouselModal
        mediaFile={selectedMedia}
        showModal={showModal}
        onClose={() => {
          setShowModal(false)
        }}
      />
      <div className={mediaList?.length > 0 ? 'd-flex w-100 gap-s4 align-items-center' : 'd-none'}>
        <Button
          onClick={jumpToLeft}
          ref={buttonLeftRef}
          variant="icon-simple"
          size="lg"
          className="gallery-arrow d-xxl-none"
          aria-label="Move to left"
        >
          <BsChevronLeft />
        </Button>
        <div id="gallery-container" className={`gallery gap-s5 ${className}`} ref={galleryRef}>
          {mediaList.map((media, index) => (
            <MediaCarouselCard key={index} mediaFile={media} index={index} onClick={onMediaClick} />
          ))}
        </div>
        <Button
          onClick={jumpToRight}
          ref={buttonRightRef}
          variant="icon-simple"
          size="lg"
          className="gallery-arrow d-xxl-none"
          aria-label="Move to right"
        >
          <BsChevronRight />
        </Button>
      </div>
    </>
  )
}

export default MediaCarousel
