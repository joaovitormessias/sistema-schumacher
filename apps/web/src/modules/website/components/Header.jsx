import { Navbar, NavbarBrand, NavbarContent, NavbarItem, NavbarMenuToggle, NavbarMenu, NavbarMenuItem, Button, Dropdown, DropdownTrigger, DropdownMenu, DropdownItem } from '@heroui/react'
import { useState } from 'react'
import { Link, useLocation, useNavigate } from 'react-router-dom'
import { ChevronDown, MapPin } from 'lucide-react'

const menuItems = [
    { name: 'Sobre', href: '#sobre', isAnchor: true },
    { name: 'Frota', href: '#frota', isAnchor: true },
    { name: 'Orçamento', href: '/orcamento', isAnchor: false },
    { name: 'Depoimentos', href: '#depoimentos', isAnchor: true },
]

const destinations = [
    { name: '🏝️ Lençóis Maranhenses', href: '/viagens/maranhao', badge: 'Mais procurado' },
    { name: '🎢 Santa Catarina', href: '/viagens/santa-catarina', badge: null },
]

export default function Header() {
    const [isMenuOpen, setIsMenuOpen] = useState(false)
    const location = useLocation()
    const navigate = useNavigate()
    const isHome = location.pathname === '/'

    const handleNavigation = (item) => {
        if (item.isAnchor) {
            if (!isHome) {
                navigate('/')
                setTimeout(() => {
                    const element = document.querySelector(item.href)
                    element?.scrollIntoView({ behavior: 'smooth' })
                }, 100)
            } else {
                const element = document.querySelector(item.href)
                element?.scrollIntoView({ behavior: 'smooth' })
            }
        } else {
            navigate(item.href)
        }
        setIsMenuOpen(false)
    }

    return (
        <Navbar
            isMenuOpen={isMenuOpen}
            onMenuOpenChange={setIsMenuOpen}
            className="bg-white/90 backdrop-blur-lg border-b border-gold-100 fixed top-0 z-50"
            maxWidth="xl"
        >
            <NavbarContent>
                <NavbarMenuToggle
                    aria-label={isMenuOpen ? "Fechar menu" : "Abrir menu"}
                    className="sm:hidden text-gold-500"
                />
                <NavbarBrand>
                    <Link to="/" className="font-heading font-bold text-2xl text-dark-900">
                        Schumacher <span className="text-gold-500">Tur</span>
                    </Link>
                </NavbarBrand>
            </NavbarContent>

            <NavbarContent className="hidden sm:flex gap-6" justify="center">
                {/* Dropdown Viagens */}
                <Dropdown
                    classNames={{
                        content: "bg-white/70 backdrop-blur-xl shadow-xl border border-white/20",
                    }}
                >
                    <NavbarItem>
                        <DropdownTrigger>
                            <Button
                                disableRipple
                                className="p-0 bg-transparent data-[hover=true]:bg-transparent text-dark-600 hover:text-gold-500 font-medium"
                                endContent={<ChevronDown size={16} />}
                                variant="light"
                            >
                                Viagens
                            </Button>
                        </DropdownTrigger>
                    </NavbarItem>
                    <DropdownMenu
                        aria-label="Destinos"
                        className="w-64"
                        itemClasses={{
                            base: [
                                "gap-4",
                                "transition-colors",
                                "data-[hover=true]:bg-gold-50/50",
                                "data-[hover=true]:text-gold-700",
                            ],
                            title: "font-semibold",
                            description: "text-gold-600/70 text-xs",
                        }}
                    >
                        {destinations.map((dest) => (
                            <DropdownItem
                                key={dest.href}
                                description={dest.badge}
                                startContent={<MapPin size={18} className="text-gold-500" />}
                                onClick={() => navigate(dest.href)}
                            >
                                {dest.name}
                            </DropdownItem>
                        ))}
                    </DropdownMenu>
                </Dropdown>

                {/* Menu Items normais */}
                {menuItems.map((item) => (
                    <NavbarItem key={item.name}>
                        <button
                            onClick={() => handleNavigation(item)}
                            className="text-dark-600 hover:text-gold-500 font-medium transition-colors duration-200"
                        >
                            {item.name}
                        </button>
                    </NavbarItem>
                ))}
            </NavbarContent>

            <NavbarContent justify="end">
                <NavbarItem>
                    <Button
                        size="sm"
                        className="bg-gold-400 text-white font-semibold px-6 hover:bg-gold-500 transition-colors"
                        onClick={() => navigate('/orcamento')}
                    >
                        Solicitar Orçamento
                    </Button>
                </NavbarItem>
            </NavbarContent>

            {/* Mobile Menu */}
            <NavbarMenu className="bg-white/95 backdrop-blur-lg pt-6">
                {/* Viagens no mobile */}
                <NavbarMenuItem>
                    <p className="text-xs uppercase tracking-wider text-dark-400 mb-2 mt-2">Viagens</p>
                </NavbarMenuItem>
                {destinations.map((dest) => (
                    <NavbarMenuItem key={dest.href}>
                        <Link
                            to={dest.href}
                            onClick={() => setIsMenuOpen(false)}
                            className="w-full text-left py-2 text-lg text-dark-700 hover:text-gold-500 font-medium transition-colors flex items-center gap-2"
                        >
                            {dest.name}
                            {dest.badge && (
                                <span className="text-xs bg-gold-100 text-gold-600 px-2 py-0.5 rounded-full">
                                    {dest.badge}
                                </span>
                            )}
                        </Link>
                    </NavbarMenuItem>
                ))}

                <NavbarMenuItem>
                    <div className="border-t border-light-200 my-4" />
                </NavbarMenuItem>

                {menuItems.map((item, index) => (
                    <NavbarMenuItem key={`${item.name}-${index}`}>
                        <button
                            onClick={() => handleNavigation(item)}
                            className="w-full text-left py-3 text-lg text-dark-700 hover:text-gold-500 font-medium transition-colors"
                        >
                            {item.name}
                        </button>
                    </NavbarMenuItem>
                ))}
            </NavbarMenu>
        </Navbar>
    )
}
