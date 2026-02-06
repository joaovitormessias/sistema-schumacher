import { motion } from 'framer-motion'
import { Card, CardBody } from '@heroui/react'
import BookingForm from '../components/BookingForm'
import { Phone, Mail, MapPin, Clock } from 'lucide-react'

export default function QuotePage() {
    return (
        <div className="bg-light-50 min-h-screen">
            {/* Hero */}
            <section className="bg-gradient-to-br from-gold-500 to-gold-600 py-16">
                <div className="container-max text-center">
                    <motion.div
                        initial={{ opacity: 0, y: 20 }}
                        animate={{ opacity: 1, y: 0 }}
                    >
                        <h1 className="text-3xl md:text-4xl font-heading font-bold text-white mb-4">
                            Solicite seu Orçamento
                        </h1>
                        <p className="text-gold-100 text-lg max-w-xl mx-auto">
                            Preencha o formulário abaixo e nossa equipe entrará em contato com a melhor proposta para sua viagem!
                        </p>
                    </motion.div>
                </div>
            </section>

            {/* Form Section */}
            <section className="section-padding -mt-8">
                <div className="container-max max-w-4xl">
                    <div className="grid md:grid-cols-3 gap-8">
                        {/* Formulário */}
                        <motion.div
                            initial={{ opacity: 0, y: 20 }}
                            animate={{ opacity: 1, y: 0 }}
                            transition={{ delay: 0.1 }}
                            className="md:col-span-2"
                        >
                            <Card className="shadow-xl">
                                <CardBody className="p-6 md:p-8">
                                    <BookingForm />
                                </CardBody>
                            </Card>
                        </motion.div>

                        {/* Sidebar */}
                        <motion.div
                            initial={{ opacity: 0, y: 20 }}
                            animate={{ opacity: 1, y: 0 }}
                            transition={{ delay: 0.2 }}
                            className="space-y-6"
                        >
                            {/* Contatos */}
                            <Card className="bg-white">
                                <CardBody className="p-5">
                                    <h3 className="font-bold text-dark-900 mb-4">Contatos Diretos</h3>
                                    <div className="space-y-4">
                                        <a
                                            href="tel:+554932466666"
                                            className="flex items-center gap-3 text-dark-600 hover:text-gold-500 transition-colors"
                                        >
                                            <div className="w-10 h-10 rounded-full bg-gold-50 flex items-center justify-center">
                                                <Phone size={18} className="text-gold-500" />
                                            </div>
                                            <div>
                                                <p className="text-sm text-dark-400">Telefone</p>
                                                <p className="font-medium">(49) 3246-6666</p>
                                            </div>
                                        </a>
                                        <a
                                            href="https://wa.me/5549999862222"
                                            target="_blank"
                                            rel="noopener noreferrer"
                                            className="flex items-center gap-3 text-dark-600 hover:text-green-500 transition-colors"
                                        >
                                            <div className="w-10 h-10 rounded-full bg-green-50 flex items-center justify-center">
                                                <Phone size={18} className="text-green-500" />
                                            </div>
                                            <div>
                                                <p className="text-sm text-dark-400">WhatsApp</p>
                                                <p className="font-medium">(49) 99986-2222</p>
                                            </div>
                                        </a>
                                        <a
                                            href="mailto:turismo@schumacher.tur.br"
                                            className="flex items-center gap-3 text-dark-600 hover:text-gold-500 transition-colors"
                                        >
                                            <div className="w-10 h-10 rounded-full bg-gold-50 flex items-center justify-center">
                                                <Mail size={18} className="text-gold-500" />
                                            </div>
                                            <div>
                                                <p className="text-sm text-dark-400">E-mail</p>
                                                <p className="font-medium text-sm">turismo@schumacher.tur.br</p>
                                            </div>
                                        </a>
                                    </div>
                                </CardBody>
                            </Card>

                            {/* Endereço */}
                            <Card className="bg-white">
                                <CardBody className="p-5">
                                    <h3 className="font-bold text-dark-900 mb-4">Nossa Sede</h3>
                                    <div className="flex items-start gap-3">
                                        <div className="w-10 h-10 rounded-full bg-gold-50 flex items-center justify-center flex-shrink-0">
                                            <MapPin size={18} className="text-gold-500" />
                                        </div>
                                        <div>
                                            <p className="text-dark-600 text-sm">
                                                SC-355, KM 35<br />
                                                Sala Comercial CONTAINER<br />
                                                São Sebastião, Fraiburgo/SC<br />
                                                CEP: 89580-000
                                            </p>
                                        </div>
                                    </div>
                                </CardBody>
                            </Card>

                            {/* Horário */}
                            <Card className="bg-white">
                                <CardBody className="p-5">
                                    <h3 className="font-bold text-dark-900 mb-4">Atendimento</h3>
                                    <div className="flex items-start gap-3">
                                        <div className="w-10 h-10 rounded-full bg-gold-50 flex items-center justify-center flex-shrink-0">
                                            <Clock size={18} className="text-gold-500" />
                                        </div>
                                        <div>
                                            <p className="text-dark-600 text-sm">
                                                <strong>Seg - Sex:</strong> 8h às 18h<br />
                                                <strong>Sábado:</strong> 8h às 12h<br />
                                                <strong>WhatsApp:</strong> 24h
                                            </p>
                                        </div>
                                    </div>
                                </CardBody>
                            </Card>
                        </motion.div>
                    </div>
                </div>
            </section>
        </div>
    )
}
