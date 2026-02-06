import { Card, CardBody } from '@heroui/react'
import { motion } from 'framer-motion'

const features = [
    {
        icon: '🛡️',
        title: 'Segurança',
        description: 'Frota com manutenção rigorosa e motoristas qualificados',
    },
    {
        icon: '⏰',
        title: 'Pontualidade',
        description: 'Compromisso com horários e planejamento de rotas',
    },
    {
        icon: '✨',
        title: 'Conforto',
        description: 'Veículos modernos equipados para sua comodidade',
    },
    {
        icon: '💬',
        title: 'Atendimento',
        description: 'Suporte dedicado do planejamento à execução',
    },
]

const containerVariants = {
    hidden: { opacity: 0 },
    visible: {
        opacity: 1,
        transition: {
            staggerChildren: 0.15,
        },
    },
}

const itemVariants = {
    hidden: { opacity: 0, y: 30 },
    visible: {
        opacity: 1,
        y: 0,
        transition: { duration: 0.6, ease: "easeOut" }
    },
}

export default function About() {
    return (
        <section id="sobre" className="section-padding bg-light-100">
            <div className="container-max">
                <motion.div
                    initial={{ opacity: 0, y: 20 }}
                    whileInView={{ opacity: 1, y: 0 }}
                    viewport={{ once: true, margin: "-100px" }}
                    transition={{ duration: 0.6 }}
                    className="text-center mb-16"
                >
                    <h2 className="section-title">
                        Sobre a <span className="text-gradient-gold">Schumacher Tur</span>
                    </h2>
                    <p className="section-subtitle">
                        A Schumacher Tur é uma empresa de turismo que atua com uma proposta diferenciada,
                        cujo foco principal não se restringe somente à venda de pacotes de viagem.
                        <span className="text-gold-500 font-semibold"> Focamos na sua experiência de viagem</span>,
                        oferecendo roteiros exclusivos para os Lençóis Maranhenses e destinos em Santa Catarina.
                    </p>
                    <p className="text-sm text-dark-400 mt-4">
                        Sede em Fraiburgo/SC • CNPJ: 17.246.217/0001-89
                    </p>
                </motion.div>

                <motion.div
                    variants={containerVariants}
                    initial="hidden"
                    whileInView="visible"
                    viewport={{ once: true, margin: "-50px" }}
                    className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6"
                >
                    {features.map((feature, index) => (
                        <motion.div key={index} variants={itemVariants}>
                            <Card className="bg-white border border-gold-100 hover:border-gold-300 hover:shadow-gold transition-all duration-300 h-full">
                                <CardBody className="text-center p-8">
                                    <div className="text-5xl mb-4">{feature.icon}</div>
                                    <h3 className="text-xl font-bold text-dark-900 mb-3">
                                        {feature.title}
                                    </h3>
                                    <p className="text-dark-500 leading-relaxed">
                                        {feature.description}
                                    </p>
                                </CardBody>
                            </Card>
                        </motion.div>
                    ))}
                </motion.div>
            </div>
        </section>
    )
}
