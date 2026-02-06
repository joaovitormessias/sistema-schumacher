import { useState } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import { X, ChevronLeft, ChevronRight, ZoomIn } from 'lucide-react'

// Imagens organizadas por categoria
const galleries = {
    maranhao: {
        title: 'Lençóis Maranhenses',
        images: [
            { src: 'https://images.unsplash.com/photo-1559128010-7c1ad6e1b6a5?w=800&q=80', caption: 'Lagoas cristalinas' },
            { src: 'https://images.unsplash.com/photo-1506905925346-21bda4d32df4?w=800&q=80', caption: 'Dunas infinitas' },
            { src: 'https://images.unsplash.com/photo-1476514525535-07fb3b4ae5f1?w=800&q=80', caption: 'Pôr do sol' },
            { src: 'https://images.unsplash.com/photo-1507525428034-b723cf961d3e?w=800&q=80', caption: 'Paisagem única' },
        ],
    },
    santaCatarina: {
        title: 'Santa Catarina',
        images: [
            { src: 'https://images.unsplash.com/photo-1600623471616-8c1966c91ff6?w=800&q=80', caption: 'Balneário Camboriú' },
            { src: 'https://images.unsplash.com/photo-1569154941061-e231b4725ef1?w=800&q=80', caption: 'Diversão em família' },
            { src: 'https://images.unsplash.com/photo-1507525428034-b723cf961d3e?w=800&q=80', caption: 'Praias paradisíacas' },
        ],
    },
    frota: {
        title: 'Nossa Frota',
        images: [
            { src: 'https://schumacher.tur.br/wp-content/uploads/2022/07/WhatsApp-Image-2022-05-19-at-11.36.15-1.jpeg', caption: 'Frota completa' },
            { src: 'https://schumacher.tur.br/wp-content/uploads/2022/07/1200.jpg', caption: 'Semi Leito Paradiso G7' },
            { src: 'https://schumacher.tur.br/wp-content/uploads/2022/07/img-e8564.jpg', caption: 'Double Decker Premium' },
            { src: 'https://schumacher.tur.br/wp-content/uploads/2022/07/img-9546.jpg', caption: 'Panorâmico 1550' },
        ],
    },
}

export default function PhotoGallery({ category = 'maranhao', showTitle = true }) {
    const [lightboxOpen, setLightboxOpen] = useState(false)
    const [currentIndex, setCurrentIndex] = useState(0)

    const gallery = galleries[category]
    if (!gallery) return null

    const openLightbox = (index) => {
        setCurrentIndex(index)
        setLightboxOpen(true)
        document.body.style.overflow = 'hidden'
    }

    const closeLightbox = () => {
        setLightboxOpen(false)
        document.body.style.overflow = 'auto'
    }

    const nextImage = () => {
        setCurrentIndex((prev) => (prev + 1) % gallery.images.length)
    }

    const prevImage = () => {
        setCurrentIndex((prev) => (prev - 1 + gallery.images.length) % gallery.images.length)
    }

    const handleKeyDown = (e) => {
        if (e.key === 'Escape') closeLightbox()
        if (e.key === 'ArrowRight') nextImage()
        if (e.key === 'ArrowLeft') prevImage()
    }

    return (
        <div>
            {showTitle && (
                <h3 className="text-xl font-bold text-dark-900 mb-4">{gallery.title}</h3>
            )}

            {/* Grid de imagens */}
            <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
                {gallery.images.map((image, index) => (
                    <motion.div
                        key={index}
                        initial={{ opacity: 0, scale: 0.9 }}
                        whileInView={{ opacity: 1, scale: 1 }}
                        viewport={{ once: true }}
                        transition={{ delay: index * 0.1 }}
                        className="relative group cursor-pointer overflow-hidden rounded-xl aspect-square"
                        onClick={() => openLightbox(index)}
                    >
                        <img
                            src={image.src}
                            alt={image.caption}
                            className="w-full h-full object-cover group-hover:scale-110 transition-transform duration-500"
                        />
                        <div className="absolute inset-0 bg-dark-900/0 group-hover:bg-dark-900/40 transition-colors flex items-center justify-center">
                            <ZoomIn size={32} className="text-white opacity-0 group-hover:opacity-100 transition-opacity" />
                        </div>
                        {image.caption && (
                            <div className="absolute bottom-0 left-0 right-0 p-2 bg-gradient-to-t from-dark-900/80 to-transparent opacity-0 group-hover:opacity-100 transition-opacity">
                                <p className="text-white text-xs font-medium truncate">{image.caption}</p>
                            </div>
                        )}
                    </motion.div>
                ))}
            </div>

            {/* Lightbox */}
            <AnimatePresence>
                {lightboxOpen && (
                    <motion.div
                        initial={{ opacity: 0 }}
                        animate={{ opacity: 1 }}
                        exit={{ opacity: 0 }}
                        className="fixed inset-0 z-[100] bg-dark-900/95 flex items-center justify-center p-4"
                        onClick={closeLightbox}
                        onKeyDown={handleKeyDown}
                        tabIndex={0}
                    >
                        {/* Close button */}
                        <button
                            onClick={closeLightbox}
                            className="absolute top-4 right-4 w-12 h-12 rounded-full bg-white/10 hover:bg-white/20 flex items-center justify-center text-white transition-colors z-10"
                        >
                            <X size={24} />
                        </button>

                        {/* Navigation */}
                        <button
                            onClick={(e) => { e.stopPropagation(); prevImage() }}
                            className="absolute left-4 w-12 h-12 rounded-full bg-white/10 hover:bg-white/20 flex items-center justify-center text-white transition-colors"
                        >
                            <ChevronLeft size={24} />
                        </button>
                        <button
                            onClick={(e) => { e.stopPropagation(); nextImage() }}
                            className="absolute right-4 w-12 h-12 rounded-full bg-white/10 hover:bg-white/20 flex items-center justify-center text-white transition-colors"
                        >
                            <ChevronRight size={24} />
                        </button>

                        {/* Image */}
                        <motion.div
                            key={currentIndex}
                            initial={{ opacity: 0, scale: 0.9 }}
                            animate={{ opacity: 1, scale: 1 }}
                            exit={{ opacity: 0, scale: 0.9 }}
                            className="max-w-5xl max-h-[80vh] relative"
                            onClick={(e) => e.stopPropagation()}
                        >
                            <img
                                src={gallery.images[currentIndex].src}
                                alt={gallery.images[currentIndex].caption}
                                className="max-w-full max-h-[80vh] object-contain rounded-lg"
                            />
                            {gallery.images[currentIndex].caption && (
                                <p className="text-center text-white mt-4 text-lg font-medium">
                                    {gallery.images[currentIndex].caption}
                                </p>
                            )}
                            <p className="text-center text-white/60 text-sm mt-2">
                                {currentIndex + 1} / {gallery.images.length}
                            </p>
                        </motion.div>
                    </motion.div>
                )}
            </AnimatePresence>
        </div>
    )
}

// Export para uso em múltiplas categorias
export function FullGallery() {
    return (
        <div className="space-y-12">
            {Object.entries(galleries).map(([key, gallery]) => (
                <div key={key}>
                    <PhotoGallery category={key} showTitle={true} />
                </div>
            ))}
        </div>
    )
}
