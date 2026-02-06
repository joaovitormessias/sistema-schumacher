import { Button } from '@heroui/react'
import { motion } from 'framer-motion'

const WHATSAPP_NUMBER = '5549999862222'
const WHATSAPP_MESSAGE = 'Olá! Gostaria de solicitar um orçamento de viagem'

export default function FinalCTA() {
    const handleWhatsApp = () => {
        window.open(`https://wa.me/${WHATSAPP_NUMBER}?text=${encodeURIComponent(WHATSAPP_MESSAGE)}`, '_blank')
    }

    return (
        <section className="section-padding bg-gradient-to-r from-gold-400 via-gold-500 to-gold-400 relative overflow-hidden">
            {/* Decorative elements */}
            <div className="absolute top-0 left-0 w-72 h-72 bg-white/10 rounded-full blur-3xl -translate-x-1/2 -translate-y-1/2" />
            <div className="absolute bottom-0 right-0 w-96 h-96 bg-white/10 rounded-full blur-3xl translate-x-1/2 translate-y-1/2" />

            <div className="container-max relative z-10">
                <motion.div
                    initial={{ opacity: 0, y: 20 }}
                    whileInView={{ opacity: 1, y: 0 }}
                    viewport={{ once: true }}
                    transition={{ duration: 0.6 }}
                    className="text-center"
                >
                    <h2 className="text-4xl sm:text-5xl font-bold text-white mb-6">
                        Pronto para Embarcar?
                    </h2>
                    <p className="text-xl text-white/90 mb-10 max-w-2xl mx-auto">
                        Solicite um orçamento agora mesmo e descubra como podemos tornar sua viagem inesquecível.
                    </p>

                    <motion.div
                        initial={{ opacity: 0, scale: 0.9 }}
                        whileInView={{ opacity: 1, scale: 1 }}
                        viewport={{ once: true }}
                        transition={{ duration: 0.4, delay: 0.2 }}
                    >
                        <Button
                            onClick={handleWhatsApp}
                            size="lg"
                            className="bg-white text-gold-600 font-bold text-lg px-12 py-8 rounded-2xl shadow-xl hover:shadow-2xl hover:scale-105 transition-all duration-300"
                        >
                            Falar no WhatsApp
                        </Button>
                    </motion.div>
                </motion.div>
            </div>
        </section>
    )
}
