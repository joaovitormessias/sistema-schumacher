import { useRef, useEffect, useState } from 'react'

export default function BusScrollAnimation() {
    const sectionRef = useRef(null)
    const canvasRef = useRef(null)
    const framesRef = useRef([])
    const [isLoading, setIsLoading] = useState(true)
    const [error, setError] = useState(null)
    const [canvasStyle, setCanvasStyle] = useState({ width: 0, height: 0 })
    const [frameSize, setFrameSize] = useState({ width: 0, height: 0 }) // Novo state para resolução interna

    // Carregar frames do WebP
    useEffect(() => {
        async function loadWebPFrames() {
            try {
                const response = await fetch('/assets/bus.webp')
                if (!response.ok) throw new Error(`HTTP ${response.status}`)

                const arrayBuffer = await response.arrayBuffer()

                if (!('ImageDecoder' in window)) {
                    throw new Error('ImageDecoder não suportado')
                }

                const decoder = new ImageDecoder({
                    data: arrayBuffer,
                    type: 'image/webp',
                })

                await decoder.decode({ frameIndex: 0 })
                const frameCount = decoder.tracks.selectedTrack.frameCount

                const decodedFrames = []
                for (let i = 0; i < frameCount; i++) {
                    const { image } = await decoder.decode({ frameIndex: i })
                    decodedFrames.push(image)
                }

                framesRef.current = decodedFrames

                // Configurar dimensões
                const firstFrame = decodedFrames[0]
                if (firstFrame) {
                    const width = firstFrame.displayWidth || firstFrame.width
                    const height = firstFrame.displayHeight || firstFrame.height

                    setFrameSize({ width, height })

                    // Função para ajustar tamanho do canvas
                    const updateCanvasSize = () => {
                        // Usar documentElement.clientWidth para excluir a barra de rolagem vertical da largura
                        const viewportW = document.documentElement.clientWidth
                        const viewportH = window.innerHeight
                        const imageRatio = width / height
                        const viewportRatio = viewportW / viewportH

                        let w, h
                        if (viewportRatio > imageRatio) {
                            // Viewport mais largo: fixar largura, cortar altura
                            w = viewportW
                            h = w / imageRatio
                        } else {
                            // Viewport mais alto: fixar altura, cortar largura
                            h = viewportH
                            w = h * imageRatio
                        }

                        setCanvasStyle({ width: w, height: h })
                    }

                    // Executar agora e no resize
                    updateCanvasSize()
                    window.addEventListener('resize', updateCanvasSize)

                    // Cleanup no unmount do componente (não no do loadWebPFrames)
                    // Obs: Como estamos dentro do useEffect de load, criar um cleanup específico seria complexo.
                    // Simplificação: vamos deixar o resize listener ativo ou mover para outro useEffect.
                    // Melhor abordagem: salvar frameSize no state (já feito) e ter um useEffect separado para o resize.
                }

                setIsLoading(false)

            } catch (err) {
                console.error(err)
                setError(err.message)
                setIsLoading(false)
            }
        }

        loadWebPFrames()
    }, [])

    // Desenhar frame inicial APÓS loading terminar e canvas estar pronto
    useEffect(() => {
        if (!isLoading && framesRef.current.length > 0 && canvasRef.current) {
            const ctx = canvasRef.current.getContext('2d')
            const canvas = canvasRef.current
            ctx.drawImage(framesRef.current[0], 0, 0, canvas.width, canvas.height)
        }
    }, [isLoading, frameSize]) // Redesenhar se tamanho mudar

    // Controle por scroll
    useEffect(() => {
        const frames = framesRef.current
        if (frames.length === 0) return

        const onScroll = () => {
            const section = sectionRef.current
            const canvas = canvasRef.current
            if (!section || !canvas) return

            const rect = section.getBoundingClientRect()

            // Calcular progresso apenas enquanto a seção está "presa" (sticky)
            // Começa quando o topo da seção atinge o topo da tela (rect.top <= 0)
            // Termina quando o fundo da seção atinge o fundo da tela
            // Distância total de scroll efetivo = altura da seção - altura da janela

            const scrollDistance = -rect.top
            const maxScroll = section.offsetHeight - window.innerHeight

            // Se maxScroll for <= 0 (seção menor que tela), evita divisão por zero
            if (maxScroll <= 0) return

            const progress = Math.max(0, Math.min(1, scrollDistance / maxScroll))

            const idx = Math.floor(progress * (frames.length - 1))
            const ctx = canvas.getContext('2d')

            // Usar dimensões internas corretas
            const frame = frames[idx]
            if (frame) {
                // Nota: reutilizamos as dimensões do canvas (canvas.width/height) configuradas anteriormente
                // para garantir que não haja distorção ou reset de estado
                ctx.drawImage(frame, 0, 0, canvas.width, canvas.height)
            }
        }

        window.addEventListener('scroll', onScroll, { passive: true })
        return () => window.removeEventListener('scroll', onScroll)
    }, [isLoading])

    return (
        <section
            ref={sectionRef}
            className="relative bg-white"
            style={{ height: '500vh' }}
        >
            <div className="sticky top-0 h-screen w-full flex items-center justify-center">

                {isLoading && (
                    <div className="animate-spin rounded-full h-16 w-16 border-t-2 border-b-2 border-gold-400" />
                )}

                <canvas
                    ref={canvasRef}
                    width={frameSize.width}   // Controlado pelo React
                    height={frameSize.height} // Controlado pelo React
                    style={{
                        display: isLoading ? 'none' : 'block',
                        width: canvasStyle.width || 'auto',
                        height: canvasStyle.height || 'auto',
                    }}
                />

                {error && (
                    <div className="text-red-500 text-center">
                        <p>Erro: {error}</p>
                        <img src="/assets/bus.webp" alt="Ônibus" className="max-h-[70vh] mt-4" />
                    </div>
                )}

                {/* Bottom gradient - transition to next section */}
                <div className="absolute inset-0 bg-gradient-to-t from-white via-transparent to-transparent pointer-events-none" />

                <div className="absolute bottom-16 left-0 right-0 text-center pointer-events-none">
                    <h2 className="text-4xl sm:text-5xl lg:text-6xl font-bold text-gradient-gold drop-shadow-lg mb-4">
                        Sua Jornada Começa Aqui
                    </h2>
                    <p className="text-gold-500 text-lg font-medium">↓ Continue descendo ↓</p>
                </div>
            </div>
        </section>
    )
}
