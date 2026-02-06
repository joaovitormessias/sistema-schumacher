import { useState } from 'react'
import { Button, Input, Select, SelectItem, Textarea, Card, CardBody, Chip } from '@heroui/react'
import { motion } from 'framer-motion'
import { MessageCircle, Send, User, Phone, Mail, MapPin, Calendar, Users, CheckCircle } from 'lucide-react'

const WHATSAPP_NUMBER = '5549999862222'

const destinations = [
    { key: 'maranhao', label: 'Lençóis Maranhenses (6-7 dias)' },
    { key: 'sc-balneario', label: 'Balneário Camboriú (3-4 dias)' },
    { key: 'sc-beto-carrero', label: 'Beto Carrero World (2 dias)' },
    { key: 'sc-combo', label: 'Combo SC Completo (5 dias)' },
    { key: 'fretamento', label: 'Fretamento Personalizado' },
    { key: 'outro', label: 'Outro destino' },
]

export default function BookingForm({ defaultDestination = '', onSuccess }) {
    const [formData, setFormData] = useState({
        name: '',
        phone: '',
        email: '',
        destination: defaultDestination,
        date: '',
        passengers: '',
        message: '',
    })
    const [isSubmitting, setIsSubmitting] = useState(false)
    const [isSubmitted, setIsSubmitted] = useState(false)

    const handleChange = (field, value) => {
        setFormData(prev => ({ ...prev, [field]: value }))
    }

    const formatWhatsAppMessage = () => {
        const destLabel = destinations.find(d => d.key === formData.destination)?.label || formData.destination
        return `Olá! Gostaria de fazer uma cotação:

📍 *Destino:* ${destLabel}
📅 *Data preferencial:* ${formData.date || 'A definir'}
👥 *Passageiros:* ${formData.passengers || 'A definir'}

👤 *Nome:* ${formData.name}
📱 *Telefone:* ${formData.phone}
📧 *Email:* ${formData.email}

💬 *Mensagem:* ${formData.message || 'Aguardo retorno!'}`
    }

    const handleSubmit = (e) => {
        e.preventDefault()
        setIsSubmitting(true)

        // Abre WhatsApp com a mensagem formatada
        const message = formatWhatsAppMessage()
        window.open(`https://wa.me/${WHATSAPP_NUMBER}?text=${encodeURIComponent(message)}`, '_blank')

        setTimeout(() => {
            setIsSubmitting(false)
            setIsSubmitted(true)
            onSuccess?.()
        }, 1000)
    }

    if (isSubmitted) {
        return (
            <motion.div
                initial={{ opacity: 0, scale: 0.9 }}
                animate={{ opacity: 1, scale: 1 }}
                className="text-center py-12"
            >
                <div className="w-20 h-20 mx-auto mb-6 rounded-full bg-green-100 flex items-center justify-center">
                    <CheckCircle size={40} className="text-green-500" />
                </div>
                <h3 className="text-2xl font-bold text-dark-900 mb-2">Mensagem Enviada!</h3>
                <p className="text-dark-500 mb-6">
                    Você será redirecionado para o WhatsApp. Nossa equipe responderá em breve!
                </p>
                <Button
                    variant="flat"
                    onClick={() => setIsSubmitted(false)}
                    className="bg-gold-50 text-gold-600"
                >
                    Enviar nova mensagem
                </Button>
            </motion.div>
        )
    }

    return (
        <motion.form
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            onSubmit={handleSubmit}
            className="space-y-8"
        >
            {/* Nome */}
            <Input
                label="Nome completo"
                labelPlacement="outside"
                variant="bordered"
                placeholder="Seu nome"
                value={formData.name}
                onChange={(e) => handleChange('name', e.target.value)}
                startContent={<User size={18} className="text-dark-400" />}
                isRequired
                classNames={{
                    inputWrapper: "border-light-300 hover:border-gold-300 focus-within:!border-gold-500",
                    label: "text-dark-700 font-medium",
                }}
            />

            {/* Telefone e Email */}
            <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
                <Input
                    label="WhatsApp"
                    labelPlacement="outside"
                    variant="bordered"
                    placeholder="(49) 99999-9999"
                    value={formData.phone}
                    onChange={(e) => handleChange('phone', e.target.value)}
                    startContent={<Phone size={18} className="text-dark-400" />}
                    isRequired
                    classNames={{
                        inputWrapper: "border-light-300 hover:border-gold-300 focus-within:!border-gold-500",
                        label: "text-dark-700 font-medium",
                    }}
                />
                <Input
                    label="E-mail"
                    labelPlacement="outside"
                    variant="bordered"
                    type="email"
                    placeholder="seu@email.com"
                    value={formData.email}
                    onChange={(e) => handleChange('email', e.target.value)}
                    startContent={<Mail size={18} className="text-dark-400" />}
                    classNames={{
                        inputWrapper: "border-light-300 hover:border-gold-300 focus-within:!border-gold-500",
                        label: "text-dark-700 font-medium",
                    }}
                />
            </div>

            {/* Destino */}
            <Select
                label="Destino desejado"
                labelPlacement="outside"
                variant="bordered"
                placeholder="Selecione o destino"
                selectedKeys={formData.destination ? [formData.destination] : []}
                onChange={(e) => handleChange('destination', e.target.value)}
                startContent={<MapPin size={18} className="text-dark-400" />}
                isRequired
                classNames={{
                    trigger: "border-light-300 hover:border-gold-300 focus-within:!border-gold-500",
                    label: "text-dark-700 font-medium",
                }}
            >
                {destinations.map((dest) => (
                    <SelectItem key={dest.key} value={dest.key}>
                        {dest.label}
                    </SelectItem>
                ))}
            </Select>

            {/* Data e Passageiros */}
            <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
                <Input
                    label="Data preferencial"
                    labelPlacement="outside"
                    variant="bordered"
                    type="date"
                    value={formData.date}
                    onChange={(e) => handleChange('date', e.target.value)}
                    startContent={<Calendar size={18} className="text-dark-400" />}
                    classNames={{
                        inputWrapper: "border-light-300 hover:border-gold-300 focus-within:!border-gold-500",
                        label: "text-dark-700 font-medium",
                    }}
                />
                <Input
                    label="Número de passageiros"
                    labelPlacement="outside"
                    variant="bordered"
                    type="number"
                    placeholder="Ex: 40"
                    min="1"
                    value={formData.passengers}
                    onChange={(e) => handleChange('passengers', e.target.value)}
                    startContent={<Users size={18} className="text-dark-400" />}
                    classNames={{
                        inputWrapper: "border-light-300 hover:border-gold-300 focus-within:!border-gold-500",
                        label: "text-dark-700 font-medium",
                    }}
                />
            </div>

            {/* Mensagem */}
            <Textarea
                label="Observações"
                labelPlacement="outside"
                variant="bordered"
                placeholder="Conte-nos mais sobre sua viagem... (opcional)"
                value={formData.message}
                onChange={(e) => handleChange('message', e.target.value)}
                minRows={3}
                classNames={{
                    inputWrapper: "border-light-300 hover:border-gold-300 focus-within:!border-gold-500",
                    label: "text-dark-700 font-medium",
                }}
            />

            {/* Submit */}
            <Button
                type="submit"
                size="lg"
                fullWidth
                isLoading={isSubmitting}
                className="bg-gradient-to-r from-gold-500 to-gold-400 text-white font-bold shadow-gold hover:shadow-gold-lg transition-transform active:scale-95 py-6"
            >
                {isSubmitting ? 'Enviando...' : (
                    <>
                        <MessageCircle size={20} className="mr-2" />
                        Enviar via WhatsApp
                    </>
                )}
            </Button>

            <p className="text-center text-xs text-dark-400">
                Ao enviar, você será redirecionado para o WhatsApp com sua mensagem pré-formatada.
            </p>
        </motion.form>
    )
}
