// Pacotes de viagem disponíveis
export const maranhaoPackages = [
    {
        id: 'maranhao-economico',
        name: 'Pacote Econômico',
        destination: 'Lençóis Maranhenses',
        duration: '6 dias / 5 noites',
        vehicle: 'Ônibus Convencional',
        vehicleCapacity: 43,
        price: 'Consultar',
        includes: [
            'Transporte ida e volta',
            'Hospedagem 3★',
            'Café da manhã',
            'Passeios inclusos',
        ],
        excludes: [
            'Almoço e jantar',
            'Bebidas',
            'Passeios opcionais',
            'Despesas pessoais',
        ],
        highlights: ['Lagoa Azul', 'Lagoa Bonita', 'Rio Preguiças', 'Atins'],
    },
    {
        id: 'maranhao-conforto',
        name: 'Pacote Conforto',
        destination: 'Lençóis Maranhenses',
        duration: '6 dias / 5 noites',
        vehicle: 'Semi Leito Panorâmico',
        vehicleCapacity: 43,
        price: 'Consultar',
        featured: true,
        badge: '⭐ Mais vendido',
        includes: [
            'Transporte premium',
            'Hospedagem 4★',
            'Café + almoço',
            'Passeios + guia',
            'Kit viagem',
        ],
        excludes: [
            'Jantar',
            'Bebidas alcoólicas',
            'Passeios opcionais',
        ],
        highlights: ['Lagoa Azul', 'Lagoa Bonita', 'Rio Preguiças', 'Atins', 'Mandacaru'],
    },
    {
        id: 'maranhao-premium',
        name: 'Pacote Premium',
        destination: 'Lençóis Maranhenses',
        duration: '7 dias / 6 noites',
        vehicle: 'Double Decker',
        vehicleCapacity: 56,
        price: 'Consultar',
        includes: [
            'Double Decker c/ Wi-Fi 4G',
            'Hospedagem 4★',
            'Pensão completa',
            'Todos os passeios',
            'Transfer privativo',
        ],
        excludes: [
            'Bebidas alcoólicas',
            'Compras pessoais',
        ],
        highlights: ['Lagoa Azul', 'Lagoa Bonita', 'Rio Preguiças', 'Atins', 'Mandacaru', 'Cardosa'],
    },
]

export const santaCatarinaPackages = [
    {
        id: 'sc-beto-carrero',
        name: 'Beto Carrero Express',
        destination: 'Beto Carrero World',
        duration: '2 dias / 1 noite',
        vehicle: 'Semi Leito',
        price: 'Consultar',
        includes: [
            'Transporte ida/volta',
            'Ingresso Beto Carrero',
            'Hospedagem',
            'Café da manhã',
        ],
    },
    {
        id: 'sc-balneario',
        name: 'Balneário Completo',
        destination: 'Balneário Camboriú',
        duration: '4 dias / 3 noites',
        vehicle: 'Ônibus Semi Leito',
        price: 'Consultar',
        featured: true,
        badge: '⭐ Ideal para grupos',
        includes: [
            'Ônibus semi-leito',
            'Hotel 3★ na Barra Sul',
            'Café incluído',
            'City tour',
        ],
    },
    {
        id: 'sc-combo',
        name: 'Combo SC Total',
        destination: 'Beto Carrero + Balneário',
        duration: '5 dias / 4 noites',
        vehicle: 'Semi Leito Panorâmico',
        price: 'Consultar',
        includes: [
            'Beto Carrero + Balneário',
            'Ingresso parque',
            'Hospedagem 4★',
            'Guia turístico',
        ],
    },
]

export const allPackages = [...maranhaoPackages, ...santaCatarinaPackages]

export function getPackageById(id) {
    return allPackages.find(pkg => pkg.id === id)
}

export function getPackagesByDestination(destination) {
    return allPackages.filter(pkg =>
        pkg.destination.toLowerCase().includes(destination.toLowerCase())
    )
}
