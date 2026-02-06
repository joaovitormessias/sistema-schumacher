import { Card, CardBody, Chip } from '@heroui/react'
import { motion } from 'framer-motion'
import { Wifi, Snowflake, Armchair, Zap, Tv, Bath, Box, Star, Users, MapPin } from 'lucide-react'

// Dados da frota (Mantidos hardcoded por enquanto, mas preparados para extração)
const vehicles = [
    {
        name: 'Convencional',
        model: 'Mercedes-Benz O500M - Mascarello Roma 350',
        capacity: 43,
        type: 'Executivo',
        image: 'https://schumacher.tur.br/wp-content/uploads/2022/07/WhatsApp-Image-2022-05-19-at-11.36.14-4-1.jpeg',
        features: ['Poltronas soft', 'Geladeira', 'Ar-condicionado', 'WC'],
    },
    {
        name: 'Semi Leito 1200',
        model: 'Mercedes-Benz O500RSD - Marcopolo Paradiso G7',
        capacity: 44,
        type: 'Semi Leito',
        image: 'https://schumacher.tur.br/wp-content/uploads/2022/07/1200.jpg',
        features: ['Poltronas semi leito', 'Ar/Calefação', 'Tomadas USB', 'WC'],
    },
    {
        name: 'Semi Leito 1050',
        model: 'Volvo B9R - Marcopolo Paradiso G7',
        capacity: 42,
        type: 'Semi Leito',
        image: 'https://schumacher.tur.br/wp-content/uploads/2022/07/whatsapp-image-2019-05-21-at-104537-2-800x600.jpg',
        features: ['Encosto de pernas', 'Ar-condicionado', 'Som', 'WC'],
    },
    {
        name: 'Panorâmico 1550',
        model: 'Scania K124 - Marcopolo Paradiso G6 1550 LD',
        capacity: 43,
        type: 'Panorâmico',
        image: 'https://schumacher.tur.br/wp-content/uploads/2022/07/img-9546.jpg',
        features: ['Vista panorâmica', 'Ar/Calefação', 'Tomadas', 'WC'],
    },
    {
        name: 'Double Decker 1800',
        model: 'Scania K400IB 4 eixos - Marcopolo 1800DD G7',
        capacity: 56,
        type: 'Double Decker',
        image: 'https://schumacher.tur.br/wp-content/uploads/2022/07/img-e8564.jpg',
        features: ['Leito premium', 'Internet 4G', 'Super soft', 'Dois andares'],
        highlight: true,
    },
    {
        name: 'Panorâmico 1550 LD',
        model: 'Mercedes-Benz O400RSD - Marcopolo Paradiso G6',
        capacity: 40,
        type: 'Panorâmico',
        image: 'https://schumacher.tur.br/wp-content/uploads/2022/07/ld-800x600.png',
        features: ['Vista panorâmica', 'Poltronas soft', 'Geladeira', 'WC'],
    },
]

const micros = [
    {
        name: 'Micro W9 Limousine',
        model: 'Volare W9 Limousine Executivo',
        capacity: 26,
        image: 'https://schumacher.tur.br/wp-content/uploads/2022/07/dsc-3565-800x600.jpg',
    },
    {
        name: 'Micro Senior G7',
        model: 'Volksbus 9.160 - Marcopolo Senior G7',
        capacity: 25,
        image: 'https://schumacher.tur.br/wp-content/uploads/2022/07/img-9556.jpg',
    },
    {
        name: 'Sprinter Executiva',
        model: 'Mercedes-Benz Sprinter 415',
        capacity: 18,
        image: 'https://schumacher.tur.br/wp-content/uploads/2022/07/whatsapp-image-2020-07-27-at-15-27-40-800x600.jpg',
    },
]

const containerVariants = {
    hidden: { opacity: 0 },
    visible: {
        opacity: 1,
        transition: { staggerChildren: 0.1 },
    },
}

const itemVariants = {
    hidden: { opacity: 0, y: 30 },
    visible: { opacity: 1, y: 0, transition: { duration: 0.5 } },
}

export default function Fleet() {
    return (
        <section id="frota" className="section-padding bg-light-50 relative overflow-hidden">
            {/* Background Elements */}
            <div className="absolute top-0 right-0 w-1/3 h-1/3 bg-gold-100/30 rounded-full blur-3xl -translate-y-1/2 translate-x-1/2" />
            <div className="absolute bottom-0 left-0 w-1/4 h-1/4 bg-gold-200/20 rounded-full blur-3xl translate-y-1/2 -translate-x-1/3" />

            <div className="container-max relative z-10">
                {/* Header */}
                <motion.div
                    initial={{ opacity: 0, y: 20 }}
                    whileInView={{ opacity: 1, y: 0 }}
                    viewport={{ once: true }}
                    transition={{ duration: 0.6 }}
                    className="text-center mb-16"
                >
                    <h2 className="section-title">
                        Nossa <span className="text-gradient-gold">Frota</span>
                    </h2>
                    <p className="section-subtitle">
                        Veículos <span className="text-gold-500 font-semibold">modernos e equipados</span> para
                        garantir o máximo conforto em suas viagens
                    </p>
                </motion.div>

                {/* Stats Bar */}
                <motion.div
                    initial={{ opacity: 0, scale: 0.95 }}
                    whileInView={{ opacity: 1, scale: 1 }}
                    viewport={{ once: true }}
                    className="relative overflow-hidden bg-white/50 backdrop-blur-md border border-gold-200 rounded-3xl p-8 mb-20 shadow-xl"
                >
                    <div className="absolute inset-0 bg-gradient-to-r from-gold-50/50 via-white/50 to-gold-50/50" />

                    <div className="relative z-10 grid grid-cols-1 md:grid-cols-3 gap-8 text-center divide-y md:divide-y-0 md:divide-x divide-gold-200">
                        <div className="group p-4">
                            <div className="mb-3 inline-flex items-center justify-center w-12 h-12 rounded-full bg-gold-100 text-gold-600 group-hover:scale-110 transition-transform duration-300">
                                <Users size={24} />
                            </div>
                            <div className="text-4xl font-heading font-bold text-transparent bg-clip-text bg-gradient-to-br from-gold-600 to-gold-400">11</div>
                            <div className="text-dark-500 font-medium mt-1">Veículos Premium</div>
                        </div>
                        <div className="group p-4">
                            <div className="mb-3 inline-flex items-center justify-center w-12 h-12 rounded-full bg-gold-100 text-gold-600 group-hover:scale-110 transition-transform duration-300">
                                <Armchair size={24} />
                            </div>
                            <div className="text-4xl font-heading font-bold text-transparent bg-clip-text bg-gradient-to-br from-gold-600 to-gold-400">380+</div>
                            <div className="text-dark-500 font-medium mt-1">Lugares Confortáveis</div>
                        </div>
                        <div className="group p-4">
                            <div className="mb-3 inline-flex items-center justify-center w-12 h-12 rounded-full bg-gold-100 text-gold-600 group-hover:scale-110 transition-transform duration-300">
                                <MapPin size={24} />
                            </div>
                            <div className="text-4xl font-heading font-bold text-transparent bg-clip-text bg-gradient-to-br from-gold-600 to-gold-400">24/7</div>
                            <div className="text-dark-500 font-medium mt-1">Suporte em Viagem</div>
                        </div>
                    </div>
                </motion.div>

                {/* Main Fleet - Ônibus */}
                <motion.div
                    variants={containerVariants}
                    initial="hidden"
                    whileInView="visible"
                    viewport={{ once: true }}
                    className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-8 mb-20"
                >
                    {vehicles.map((vehicle, index) => (
                        <motion.div key={index} variants={itemVariants}>
                            <Card className={`overflow-hidden h-full group bg-white ${vehicle.highlight ? 'ring-2 ring-gold-400 shadow-gold-lg' : 'border border-light-200 hover:border-gold-300'} shadow-sm hover:shadow-xl transition-all duration-500`}>
                                <div className="relative h-64 overflow-hidden">
                                    <div className="absolute inset-0 bg-dark-900/10 group-hover:bg-transparent transition-colors duration-500 z-10" />
                                    <img
                                        src={vehicle.image}
                                        alt={vehicle.name}
                                        className="w-full h-full object-cover group-hover:scale-110 transition-transform duration-700"
                                    />
                                    {vehicle.highlight && (
                                        <div className="absolute top-4 right-4 z-20 bg-gradient-to-r from-gold-500 to-gold-400 text-white px-3 py-1.5 rounded-full text-xs font-bold shadow-lg flex items-center gap-1">
                                            <Star size={12} fill="currentColor" /> PREMIUM
                                        </div>
                                    )}
                                    <div className="absolute bottom-0 left-0 right-0 z-20 p-4 bg-gradient-to-t from-dark-900/90 via-dark-900/60 to-transparent">
                                        <div className="flex items-center justify-between">
                                            <Chip size="sm" className="bg-gold-500 text-white font-semibold">
                                                {vehicle.capacity} lugares
                                            </Chip>
                                            <span className="text-white/90 text-sm font-medium backdrop-blur-sm bg-white/10 px-2 py-1 rounded">
                                                {vehicle.type}
                                            </span>
                                        </div>
                                    </div>
                                </div>
                                <CardBody className="p-6">
                                    <h3 className="font-heading font-bold text-xl text-dark-900 mb-2 group-hover:text-gold-600 transition-colors">
                                        {vehicle.name}
                                    </h3>
                                    <p className="text-sm text-dark-400 mb-4 pb-4 border-b border-light-200">
                                        {vehicle.model}
                                    </p>
                                    <div className="flex flex-wrap gap-2">
                                        {vehicle.features.map((feature, idx) => (
                                            <span key={idx} className="inline-flex items-center gap-1.5 text-xs font-medium bg-light-100 text-dark-600 px-2.5 py-1 rounded-full border border-light-200">
                                                <FeatureIcon feature={feature} />
                                                {feature}
                                            </span>
                                        ))}
                                    </div>
                                </CardBody>
                            </Card>
                        </motion.div>
                    ))}
                </motion.div>

                {/* Micro e Vans */}
                <motion.div
                    initial={{ opacity: 0 }}
                    whileInView={{ opacity: 1 }}
                    viewport={{ once: true }}
                    className="mb-12"
                >
                    <div className="flex items-center gap-4 mb-8">
                        <div className="h-px bg-gradient-to-r from-transparent via-gold-200 to-transparent flex-1" />
                        <div className="text-center">
                            <h3 className="text-2xl font-bold text-dark-800">
                                Micro-ônibus e Vans
                            </h3>
                            <p className="text-dark-400">Opções versáteis para grupos menores</p>
                        </div>
                        <div className="h-px bg-gradient-to-r from-transparent via-gold-200 to-transparent flex-1" />
                    </div>
                </motion.div>

                <motion.div
                    variants={containerVariants}
                    initial="hidden"
                    whileInView="visible"
                    viewport={{ once: true }}
                    className="grid grid-cols-1 sm:grid-cols-3 gap-6"
                >
                    {micros.map((micro, index) => (
                        <motion.div key={index} variants={itemVariants}>
                            <Card className="overflow-hidden bg-white border border-light-200 hover:border-gold-300 shadow-sm hover:shadow-lg transition-all duration-300 group">
                                <div className="relative h-48 overflow-hidden">
                                    <img
                                        src={micro.image}
                                        alt={micro.name}
                                        className="w-full h-full object-cover group-hover:scale-105 transition-transform duration-500"
                                    />
                                    <div className="absolute bottom-3 right-3">
                                        <Chip size="sm" className="bg-dark-900/80 backdrop-blur-md text-white border border-white/20">
                                            {micro.capacity} lugares
                                        </Chip>
                                    </div>
                                </div>
                                <CardBody className="p-5">
                                    <h4 className="font-heading font-bold text-dark-900 text-lg mb-1 group-hover:text-gold-600 transition-colors">
                                        {micro.name}
                                    </h4>
                                    <p className="text-xs text-dark-400">{micro.model}</p>
                                </CardBody>
                            </Card>
                        </motion.div>
                    ))}
                </motion.div>

                {/* Amenities Legend */}
                <motion.div
                    initial={{ opacity: 0, y: 20 }}
                    whileInView={{ opacity: 1, y: 0 }}
                    viewport={{ once: true }}
                    className="mt-20 pt-10 border-t border-gold-100"
                >
                    <div className="text-center mb-8">
                        <span className="inline-block px-4 py-1 rounded-full bg-gold-50 text-gold-600 text-xs font-bold tracking-wider uppercase mb-2">
                            Conforto a bordo
                        </span>
                        <h4 className="text-xl font-bold text-dark-800">Tudo o que você precisa para uma viagem perfeita</h4>
                    </div>

                    <div className="max-w-4xl mx-auto flex flex-wrap justify-center gap-8 md:gap-12">
                        <AmenityItem icon={Snowflake} label="Ar Climatizado" />
                        <AmenityItem icon={Wifi} label="Wi-Fi 4G" />
                        <AmenityItem icon={Armchair} label="Poltronas Soft" />
                        <AmenityItem icon={Zap} label="Carregadores USB" />
                        <AmenityItem icon={Tv} label="Multimídia" />
                        <AmenityItem icon={Bath} label="Toilette" />
                        <AmenityItem icon={Box} label="Geladeira" />
                    </div>
                    <p className="text-center text-xs text-dark-400 mt-8 opacity-60">
                        * A disponibilidade dos itens varia conforme a configuração de cada veículo.
                    </p>
                </motion.div>
            </div>
        </section>
    )
}

function FeatureIcon({ feature }) {
    const text = feature.toLowerCase()
    if (text.includes('ar')) return <Snowflake size={14} className="text-gold-500" />
    if (text.includes('wifi') || text.includes('internet')) return <Wifi size={14} className="text-gold-500" />
    if (text.includes('poltrona') || text.includes('leito') || text.includes('encosto')) return <Armchair size={14} className="text-gold-500" />
    if (text.includes('usb') || text.includes('tomadas')) return <Zap size={14} className="text-gold-500" />
    if (text.includes('som') || text.includes('tv')) return <Tv size={14} className="text-gold-500" />
    if (text.includes('wc') || text.includes('banheiro')) return <Bath size={14} className="text-gold-500" />
    if (text.includes('geladeira')) return <Box size={14} className="text-gold-500" />
    return <Star size={14} className="text-gold-500" />
}

function AmenityItem({ icon: Icon, label }) {
    return (
        <div className="flex flex-col items-center gap-2 group cursor-default">
            <div className="w-12 h-12 rounded-2xl bg-white border border-light-200 shadow-sm flex items-center justify-center text-dark-400 group-hover:text-gold-500 group-hover:border-gold-300 group-hover:shadow-gold transition-all duration-300">
                <Icon size={24} strokeWidth={1.5} />
            </div>
            <span className="text-xs font-medium text-dark-500 group-hover:text-dark-900 transition-colors">
                {label}
            </span>
        </div>
    )
}
