import { Divider } from '@heroui/react'
import { Link, useNavigate } from 'react-router-dom'

const links = {
    empresa: [
        { name: 'Sobre', href: '#sobre' },
        { name: 'Serviços', href: '#servicos' },
        { name: 'Nossa Frota', href: '#frota' },
        { name: 'Depoimentos', href: '#depoimentos' },
        { name: 'Solicitar Orçamento', href: '/orcamento', isRoute: true },
    ],
    contato: [
        { name: '📞 (49) 3246-6666', href: 'tel:+554932466666' },
        { name: '💬 (49) 99986-2222', href: 'https://wa.me/5549999862222' },
        { name: '📧 turismo@schumacher.tur.br', href: 'mailto:turismo@schumacher.tur.br' },
        { name: '📍 SC-355, KM 35 - Fraiburgo/SC', href: 'https://maps.google.com/?q=SC-355+KM+35+Fraiburgo+SC' },
    ],
}

export default function Footer() {
    const navigate = useNavigate()

    const handleNavigation = (link) => {
        if (link.isRoute) {
            navigate(link.href)
            window.scrollTo({ top: 0, behavior: 'smooth' })
        } else if (link.href.startsWith('#')) {
            if (window.location.pathname !== '/') {
                navigate('/')
                setTimeout(() => {
                    document.querySelector(link.href)?.scrollIntoView({ behavior: 'smooth' })
                }, 100)
            } else {
                document.querySelector(link.href)?.scrollIntoView({ behavior: 'smooth' })
            }
        } else {
            window.open(link.href, '_blank')
        }
    }

    return (
        <footer id="contato" className="bg-light-200 border-t border-light-300">
            <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-16">
                <div className="grid grid-cols-1 md:grid-cols-3 gap-12">
                    {/* Brand */}
                    <div>
                        <h3 className="font-heading text-2xl font-bold text-dark-900 mb-4">
                            Schumacher <span className="text-gold-500">Tur</span>
                        </h3>
                        <p className="text-dark-500 leading-relaxed mb-4">
                            Sua viagem com conforto, segurança e pontualidade.
                            Especialistas em viagens ao Maranhão e turismo em Santa Catarina.
                        </p>
                        <p className="text-xs text-dark-400 mb-4">
                            CNPJ: 17.246.217/0001-89
                        </p>
                        <div className="flex gap-4">
                            <a href="https://www.facebook.com/joseane.schumachertur" target="_blank" rel="noopener noreferrer" className="w-10 h-10 rounded-full bg-gold-100 flex items-center justify-center text-gold-600 hover:bg-gold-200 transition-colors">
                                📘
                            </a>
                            <a href="https://www.instagram.com/schumacher_tur/" target="_blank" rel="noopener noreferrer" className="w-10 h-10 rounded-full bg-gold-100 flex items-center justify-center text-gold-600 hover:bg-gold-200 transition-colors">
                                📸
                            </a>
                            <a href="https://www.youtube.com/channel/UCZV5YZpjW_7QtGuHre6Tj8w" target="_blank" rel="noopener noreferrer" className="w-10 h-10 rounded-full bg-gold-100 flex items-center justify-center text-gold-600 hover:bg-gold-200 transition-colors">
                                ▶️
                            </a>
                        </div>
                    </div>

                    {/* Links */}
                    <div>
                        <h4 className="font-bold text-dark-900 mb-4">Navegação</h4>
                        <ul className="space-y-3">
                            {links.empresa.map((link) => (
                                <li key={link.name}>
                                    <button
                                        onClick={() => handleNavigation(link)}
                                        className="text-dark-500 hover:text-gold-500 transition-colors"
                                    >
                                        {link.name}
                                    </button>
                                </li>
                            ))}
                        </ul>
                    </div>

                    {/* Contact */}
                    <div>
                        <h4 className="font-bold text-dark-900 mb-4">Contato</h4>
                        <ul className="space-y-3">
                            {links.contato.map((link) => (
                                <li key={link.name}>
                                    <a
                                        href={link.href}
                                        className="text-dark-500 hover:text-gold-500 transition-colors"
                                    >
                                        {link.name}
                                    </a>
                                </li>
                            ))}
                        </ul>
                    </div>
                </div>

                <Divider className="my-8 bg-light-400" />

                <div className="flex flex-col sm:flex-row justify-between items-center gap-4">
                    <p className="text-dark-400 text-sm">
                        © {new Date().getFullYear()} Schumacher Tur. Todos os direitos reservados.
                    </p>
                    <p className="text-dark-400 text-sm">
                        Feito com 💛 em Santa Catarina
                    </p>
                </div>
            </div>
        </footer>
    )
}
