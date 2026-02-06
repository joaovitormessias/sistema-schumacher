import { Card, CardBody, Avatar } from '@heroui/react'
import { motion } from 'framer-motion'

const testimonials = [
    {
        name: 'Darlene Xavier',
        role: 'Viagem ao Maranhão',
        company: 'Lençóis Maranhenses',
        quote: 'Gostaria de agradecer por tudo, principalmente pela paciência. A viagem foi perfeita… Receptivo maravilhoso, motorista maravilhoso, guia maravilhoso, funcionários e serviços perfeitos. Sem contar o lugar visitado que é simplesmente magnífico e mágico.',
        avatar: '👩',
    },
    {
        name: 'Carlos Alberto',
        role: 'Viagem ao Maranhão',
        company: 'Experiência Inesquecível',
        quote: 'A experiência no Maranhão foi fantástica e completamente impactante. Os lugares são incríveis e deslumbrantes. Vale lembrar que é um roteiro simples mas magnífico, cheio de natureza e interação com o ambiente! As pessoas são acolhedoras e fazem você se sentir em casa, literalmente.',
        avatar: '👨',
    },
    {
        name: 'Família Campos',
        role: 'Turismo em Grupo',
        company: 'Viagem em Família',
        quote: 'A viagem foi maravilhosa. Além das belezas naturais, tiro o chapéu para a organização. Tudo feito com muito profissionalismo, pensando nos mínimos detalhes. Super recomendo.',
        avatar: '👨‍👩‍👧',
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

export default function Testimonials() {
    return (
        <section id="depoimentos" className="section-padding bg-white">
            <div className="container-max">
                <motion.div
                    initial={{ opacity: 0, y: 20 }}
                    whileInView={{ opacity: 1, y: 0 }}
                    viewport={{ once: true }}
                    transition={{ duration: 0.6 }}
                    className="text-center mb-16"
                >
                    <h2 className="section-title">
                        Quem Viaja, <span className="text-gradient-gold">Recomenda</span>
                    </h2>
                    <p className="section-subtitle">
                        Veja o que nossos clientes dizem sobre nós
                    </p>
                </motion.div>

                <motion.div
                    variants={containerVariants}
                    initial="hidden"
                    whileInView="visible"
                    viewport={{ once: true }}
                    className="grid grid-cols-1 md:grid-cols-3 gap-8"
                >
                    {testimonials.map((testimonial, index) => (
                        <motion.div key={index} variants={itemVariants}>
                            <Card className="bg-light-100 border border-light-300 hover:border-gold-300 hover:shadow-gold transition-all duration-300 h-full">
                                <CardBody className="p-8">
                                    {/* Stars */}
                                    <div className="flex gap-1 mb-4">
                                        {[...Array(5)].map((_, i) => (
                                            <span key={i} className="text-gold-400 text-xl">★</span>
                                        ))}
                                    </div>

                                    {/* Quote */}
                                    <p className="text-dark-600 leading-relaxed mb-6 italic">
                                        "{testimonial.quote}"
                                    </p>

                                    {/* Author */}
                                    <div className="flex items-center gap-4">
                                        <div className="w-12 h-12 rounded-full bg-gold-100 flex items-center justify-center text-2xl border-2 border-gold-300">
                                            {testimonial.avatar}
                                        </div>
                                        <div>
                                            <div className="font-bold text-dark-900">
                                                {testimonial.name}
                                            </div>
                                            <div className="text-sm text-dark-500">
                                                {testimonial.role}
                                            </div>
                                            <div className="text-sm text-gold-500 font-medium">
                                                {testimonial.company}
                                            </div>
                                        </div>
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
