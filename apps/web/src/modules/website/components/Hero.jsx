import { Button } from '@heroui/react'
import { motion } from 'framer-motion'

const WHATSAPP_NUMBER = '5549999862222'
const WHATSAPP_MESSAGE = 'Olá! Gostaria de informações sobre viagens com a Schumacher Tur'

// Floating particles animation
const FloatingParticle = ({ delay, duration, x, y, size }) => (
    <motion.div
        className="absolute rounded-full bg-gold-400/20"
        style={{ width: size, height: size, left: x, top: y }}
        animate={{
            y: [0, -30, 0],
            x: [0, 15, 0],
            opacity: [0.2, 0.5, 0.2],
            scale: [1, 1.2, 1],
        }}
        transition={{
            duration,
            delay,
            repeat: Infinity,
            ease: "easeInOut",
        }}
    />
)



export default function Hero() {
    const handleWhatsApp = () => {
        window.open(`https://wa.me/${WHATSAPP_NUMBER}?text=${encodeURIComponent(WHATSAPP_MESSAGE)}`, '_blank')
    }

    const scrollToFleet = () => {
        document.getElementById('frota')?.scrollIntoView({ behavior: 'smooth' })
    }

    return (
        <section className="relative min-h-screen flex items-center justify-center overflow-hidden z-10 bg-white">
            {/* Gradient background - ends in pure white for smooth transition */}
            <div className="absolute inset-0 bg-gradient-to-b from-gold-50 via-white to-white" />

            {/* Animated mesh gradient overlay */}
            <motion.div
                className="absolute inset-0 opacity-50"
                style={{
                    background: 'radial-gradient(ellipse at 20% 30%, rgba(212,175,55,0.15) 0%, transparent 50%), radial-gradient(ellipse at 80% 70%, rgba(212,175,55,0.1) 0%, transparent 50%)',
                }}
                animate={{
                    scale: [1, 1.1, 1],
                    opacity: [0.5, 0.7, 0.5],
                }}
                transition={{ duration: 8, repeat: Infinity, ease: "easeInOut" }}
            />

            {/* Floating particles */}
            <FloatingParticle delay={0} duration={4} x="10%" y="20%" size={20} />
            <FloatingParticle delay={1} duration={5} x="85%" y="15%" size={30} />
            <FloatingParticle delay={2} duration={6} x="70%" y="60%" size={25} />
            <FloatingParticle delay={0.5} duration={4.5} x="15%" y="70%" size={35} />
            <FloatingParticle delay={1.5} duration={5.5} x="50%" y="80%" size={20} />
            <FloatingParticle delay={2.5} duration={7} x="30%" y="10%" size={15} />

            {/* Decorative rings */}
            <motion.div
                className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[600px] h-[600px] border border-gold-200/30 rounded-full"
                animate={{ rotate: 360, scale: [1, 1.05, 1] }}
                transition={{ duration: 30, repeat: Infinity, ease: "linear" }}
            />
            <motion.div
                className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[800px] h-[800px] border border-gold-200/20 rounded-full"
                animate={{ rotate: -360 }}
                transition={{ duration: 40, repeat: Infinity, ease: "linear" }}
            />

            {/* Glowing orbs */}
            <div className="absolute top-20 right-[15%] w-64 h-64 bg-gradient-to-br from-gold-300/30 to-gold-400/10 rounded-full blur-3xl" />
            <div className="absolute bottom-20 left-[10%] w-80 h-80 bg-gradient-to-tr from-gold-200/25 to-transparent rounded-full blur-3xl" />

            {/* Content */}
            <div className="relative z-10 max-w-6xl mx-auto px-4 sm:px-6 lg:px-8 py-20">
                <div className="text-center">
                    {/* Badge */}
                    <motion.div
                        initial={{ opacity: 0, y: 20 }}
                        animate={{ opacity: 1, y: 0 }}
                        transition={{ duration: 0.6 }}
                        className="inline-flex items-center gap-2 bg-white/80 backdrop-blur-sm border border-gold-200 rounded-full px-4 py-2 mb-8 shadow-soft"
                    >
                        <span className="w-2 h-2 bg-green-500 rounded-full animate-pulse" />
                        <span className="text-sm font-medium text-dark-600">Atendimento 24h via WhatsApp</span>
                    </motion.div>

                    {/* Main heading */}
                    <motion.h1
                        initial={{ opacity: 0, y: 30 }}
                        animate={{ opacity: 1, y: 0 }}
                        transition={{ duration: 0.8, delay: 0.2 }}
                        className="text-5xl sm:text-6xl lg:text-7xl xl:text-8xl font-bold mb-6 leading-[1.1] text-dark-900"
                    >
                        Viagens com
                        <br />
                        <span className="relative inline-block">
                            <span className="text-gradient-gold">Conforto Premium</span>
                            <motion.span
                                className="absolute -bottom-2 left-0 right-0 h-3 bg-gold-400/20 rounded-full -z-10"
                                initial={{ scaleX: 0 }}
                                animate={{ scaleX: 1 }}
                                transition={{ duration: 0.8, delay: 0.8 }}
                            />
                        </span>
                    </motion.h1>

                    {/* Subtitle */}
                    <motion.p
                        initial={{ opacity: 0, y: 20 }}
                        animate={{ opacity: 1, y: 0 }}
                        transition={{ duration: 0.8, delay: 0.4 }}
                        className="text-lg sm:text-xl lg:text-2xl text-dark-500 mb-10 max-w-2xl mx-auto leading-relaxed"
                    >
                        Fretamento executivo, turismo e eventos com
                        <span className="text-gold-500 font-semibold"> veículos modernos </span>
                        e atendimento personalizado
                    </motion.p>

                    {/* CTA Buttons */}
                    <motion.div
                        initial={{ opacity: 0, y: 20 }}
                        animate={{ opacity: 1, y: 0 }}
                        transition={{ duration: 0.8, delay: 0.6 }}
                        className="flex flex-col sm:flex-row gap-4 justify-center items-center mb-12"
                    >
                        <motion.div whileHover={{ scale: 1.05 }} whileTap={{ scale: 0.98 }}>
                            <Button
                                onClick={handleWhatsApp}
                                size="lg"
                                className="bg-gradient-to-r from-gold-500 via-gold-400 to-gold-500 text-white font-bold text-lg px-10 py-7 rounded-2xl shadow-gold-lg hover:shadow-2xl transition-shadow duration-300 flex items-center gap-3"
                            >
                                <span className="text-2xl">💬</span>
                                Pedir Cotação Grátis
                            </Button>
                        </motion.div>

                        <motion.div whileHover={{ scale: 1.05 }} whileTap={{ scale: 0.98 }}>
                            <Button
                                onClick={scrollToFleet}
                                size="lg"
                                variant="bordered"
                                className="bg-white/50 backdrop-blur-sm border-2 border-gold-300 text-gold-600 font-semibold text-lg px-10 py-7 rounded-2xl hover:bg-gold-50 transition-all duration-300 flex items-center gap-3"
                            >
                                <span className="text-2xl">🚌</span>
                                Conhecer Frota
                            </Button>
                        </motion.div>
                    </motion.div>

                    {/* Trust badges */}
                    <motion.div
                        initial={{ opacity: 0 }}
                        animate={{ opacity: 1 }}
                        transition={{ duration: 0.8, delay: 0.8 }}
                        className="flex flex-wrap justify-center gap-6 text-dark-400"
                    >
                        <div className="flex items-center gap-2">
                            <span className="text-gold-500">✓</span>
                            <span className="text-sm">+10 anos no mercado</span>
                        </div>
                        <div className="flex items-center gap-2">
                            <span className="text-gold-500">✓</span>
                            <span className="text-sm">Frota 100% revisada</span>
                        </div>
                        <div className="flex items-center gap-2">
                            <span className="text-gold-500">✓</span>
                            <span className="text-sm">Seguro incluso</span>
                        </div>
                    </motion.div>
                </div>
            </div>

            {/* Scroll indicator */}
            <motion.div
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
                transition={{ delay: 1.2 }}
                className="absolute bottom-28 left-1/2 -translate-x-1/2 z-20"
            >
                <motion.div
                    animate={{ y: [0, 12, 0] }}
                    transition={{ duration: 1.5, repeat: Infinity, ease: "easeInOut" }}
                    className="flex flex-col items-center gap-2"
                >
                    <span className="text-xs text-dark-400 uppercase tracking-widest">Scroll</span>
                    <div className="w-6 h-10 border-2 border-gold-300 rounded-full flex items-start justify-center p-1.5">
                        <motion.div
                            animate={{ y: [0, 12, 0] }}
                            transition={{ duration: 1.5, repeat: Infinity, ease: "easeInOut" }}
                            className="w-1.5 h-3 bg-gold-400 rounded-full"
                        />
                    </div>
                </motion.div>
            </motion.div>

            {/* Smooth transition gradient to video section */}
            <div className="absolute bottom-0 left-0 right-0 h-24 bg-gradient-to-b from-transparent to-white pointer-events-none" />
        </section>
    )
}
