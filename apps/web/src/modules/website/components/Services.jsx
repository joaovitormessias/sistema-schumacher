import { Card, CardBody, CardHeader, Chip } from '@heroui/react'
import { motion } from 'framer-motion'

// Destinos em destaque (prioridade)
const destinations = [
    {
        icon: '🏝️',
        title: 'Lençóis Maranhenses',
        description: 'Viagem completa aos Lençóis Maranhenses com roteiro exclusivo. Lagoas cristalinas, dunas infinitas e paisagens de tirar o fôlego.',
        features: ['Roteiro completo', 'Hospedagem inclusa', 'Guia especializado'],
        highlight: true,
        badge: '⭐ MAIS PROCURADO',
    },
    {
        icon: '🎢',
        title: 'Santa Catarina',
        description: 'Balneário Camboriú, Beto Carrero World e praias incríveis. Diversão garantida para toda a família.',
        features: ['Beto Carrero', 'Balneário Camboriú', 'Praias paradisíacas'],
        highlight: false,
        badge: '🌴 LAZER',
    },
]

// Outros serviços
const services = [
    {
        icon: '🏢',
        title: 'Fretamento Empresarial',
        description: 'Transporte regular de colaboradores com rotas personalizadas e pontualidade garantida.',
        features: ['Rotas customizadas', 'Contratos flexíveis'],
    },
    {
        icon: '🎉',
        title: 'Eventos e Transfers',
        description: 'Transporte para eventos corporativos, casamentos, formaturas e ocasiões especiais.',
        features: ['Logística completa', 'Atendimento VIP'],
    },
]

const containerVariants = {
    hidden: { opacity: 0 },
    visible: {
        opacity: 1,
        transition: { staggerChildren: 0.15 },
    },
}

const itemVariants = {
    hidden: { opacity: 0, y: 30 },
    visible: { opacity: 1, y: 0, transition: { duration: 0.6 } },
}

export default function Services() {
    return (
        <section id="servicos" className="section-padding bg-white">
            <div className="container-max">
                <motion.div
                    initial={{ opacity: 0, y: 20 }}
                    whileInView={{ opacity: 1, y: 0 }}
                    viewport={{ once: true }}
                    transition={{ duration: 0.6 }}
                    className="text-center mb-16"
                >
                    <h2 className="section-title">
                        Nossos <span className="text-gradient-gold">Destinos</span>
                    </h2>
                    <p className="section-subtitle">
                        Viagens inesquecíveis com conforto e segurança
                    </p>
                </motion.div>

                {/* Destinos em Destaque */}
                <motion.div
                    variants={containerVariants}
                    initial="hidden"
                    whileInView="visible"
                    viewport={{ once: true }}
                    className="grid grid-cols-1 md:grid-cols-2 gap-8 mb-16"
                >
                    {destinations.map((dest, index) => (
                        <motion.div key={index} variants={itemVariants}>
                            <Card className={`border-2 ${dest.highlight ? 'border-gold-400 shadow-gold-lg' : 'border-light-300'} hover:border-gold-400 hover:shadow-gold transition-all duration-300 h-full bg-gradient-to-br from-white to-gold-50`}>
                                <CardHeader className="flex gap-4 pb-0 pt-6 px-6">
                                    <div className="text-5xl">{dest.icon}</div>
                                    <div>
                                        <Chip size="sm" className={`${dest.highlight ? 'bg-gold-500 text-white' : 'bg-gold-100 text-gold-700'} mb-2`}>
                                            {dest.badge}
                                        </Chip>
                                        <h3 className="text-2xl font-bold text-dark-900">
                                            {dest.title}
                                        </h3>
                                    </div>
                                </CardHeader>
                                <CardBody className="px-6 pb-6">
                                    <p className="text-dark-500 mb-4 leading-relaxed">
                                        {dest.description}
                                    </p>
                                    <div className="flex flex-wrap gap-2">
                                        {dest.features.map((feature, idx) => (
                                            <Chip
                                                key={idx}
                                                size="sm"
                                                variant="flat"
                                                className="bg-gold-50 text-gold-700 border border-gold-200"
                                            >
                                                ✓ {feature}
                                            </Chip>
                                        ))}
                                    </div>
                                </CardBody>
                            </Card>
                        </motion.div>
                    ))}
                </motion.div>

                {/* Outros Serviços */}
                <motion.div
                    initial={{ opacity: 0, y: 20 }}
                    whileInView={{ opacity: 1, y: 0 }}
                    viewport={{ once: true }}
                    className="text-center mb-8"
                >
                    <h3 className="text-2xl font-bold text-dark-700">
                        Também oferecemos
                    </h3>
                </motion.div>

                <motion.div
                    variants={containerVariants}
                    initial="hidden"
                    whileInView="visible"
                    viewport={{ once: true }}
                    className="grid grid-cols-1 md:grid-cols-2 gap-6"
                >
                    {services.map((service, index) => (
                        <motion.div key={index} variants={itemVariants}>
                            <Card className="bg-white border border-light-300 hover:border-gold-300 hover:shadow-gold transition-all duration-300 h-full">
                                <CardHeader className="flex gap-4 pb-0 pt-5 px-5">
                                    <div className="text-3xl">{service.icon}</div>
                                    <h3 className="text-xl font-bold text-dark-900">
                                        {service.title}
                                    </h3>
                                </CardHeader>
                                <CardBody className="px-5 pb-5">
                                    <p className="text-dark-500 mb-3 leading-relaxed text-sm">
                                        {service.description}
                                    </p>
                                    <div className="flex flex-wrap gap-2">
                                        {service.features.map((feature, idx) => (
                                            <Chip
                                                key={idx}
                                                size="sm"
                                                variant="flat"
                                                className="bg-light-200 text-dark-600 border border-light-300"
                                            >
                                                ✓ {feature}
                                            </Chip>
                                        ))}
                                    </div>
                                </CardBody>
                            </Card>
                        </motion.div>
                    ))}
                </motion.div>
            </div>
        </section>
    )
}
