// Dados da frota com imagens reais
export const buses = [
    {
        id: 'convencional',
        name: 'Convencional',
        model: 'Mercedes-Benz O500M - Mascarello Roma 350',
        capacity: 43,
        type: 'Executivo',
        image: 'https://schumacher.tur.br/wp-content/uploads/2022/07/WhatsApp-Image-2022-05-19-at-11.36.14-4-1.jpeg',
        features: ['Poltronas soft', 'Geladeira', 'Ar-condicionado', 'WC'],
        description: 'Ideal para viagens curtas e médias com conforto.',
    },
    {
        id: 'semi-leito-1200',
        name: 'Semi Leito 1200',
        model: 'Mercedes-Benz O500RSD - Marcopolo Paradiso G7',
        capacity: 44,
        type: 'Semi Leito',
        image: 'https://schumacher.tur.br/wp-content/uploads/2022/07/1200.jpg',
        features: ['Poltronas semi leito', 'Ar/Calefação', 'Tomadas USB', 'WC'],
        description: 'Conforto extra para viagens longas.',
    },
    {
        id: 'semi-leito-1050',
        name: 'Semi Leito 1050',
        model: 'Volvo B9R - Marcopolo Paradiso G7',
        capacity: 42,
        type: 'Semi Leito',
        image: 'https://schumacher.tur.br/wp-content/uploads/2022/07/whatsapp-image-2019-05-21-at-104537-2-800x600.jpg',
        features: ['Encosto de pernas', 'Ar-condicionado', 'Som', 'WC'],
        description: 'Motor Volvo com excelente desempenho.',
    },
    {
        id: 'panoramico-1550',
        name: 'Panorâmico 1550',
        model: 'Scania K124 - Marcopolo Paradiso G6 1550 LD',
        capacity: 43,
        type: 'Panorâmico',
        image: 'https://schumacher.tur.br/wp-content/uploads/2022/07/img-9546.jpg',
        features: ['Vista panorâmica', 'Ar/Calefação', 'Tomadas', 'WC'],
        description: 'Visão privilegiada das paisagens durante a viagem.',
    },
    {
        id: 'double-decker',
        name: 'Double Decker 1800',
        model: 'Scania K400IB 4 eixos - Marcopolo 1800DD G7',
        capacity: 56,
        type: 'Double Decker',
        image: 'https://schumacher.tur.br/wp-content/uploads/2022/07/img-e8564.jpg',
        features: ['Leito premium', 'Internet 4G', 'Super soft', 'Dois andares'],
        description: 'O melhor em conforto para viagens longas. Premium!',
        highlight: true,
    },
    {
        id: 'panoramico-1550-ld',
        name: 'Panorâmico 1550 LD',
        model: 'Mercedes-Benz O400RSD - Marcopolo Paradiso G6',
        capacity: 40,
        type: 'Panorâmico',
        image: 'https://schumacher.tur.br/wp-content/uploads/2022/07/ld-800x600.png',
        features: ['Vista panorâmica', 'Poltronas soft', 'Geladeira', 'WC'],
        description: 'Conforto Mercedes com vista privilegiada.',
    },
]

export const micros = [
    {
        id: 'w9-limousine',
        name: 'Micro W9 Limousine',
        model: 'Volare W9 Limousine Executivo',
        capacity: 26,
        type: 'Micro',
        image: 'https://schumacher.tur.br/wp-content/uploads/2022/07/dsc-3565-800x600.jpg',
        features: ['Poltronas soft', 'Som', 'Tomadas', 'Geladeira'],
    },
    {
        id: 'senior-g7',
        name: 'Micro Senior G7',
        model: 'Volksbus 9.160 - Marcopolo Senior G7',
        capacity: 25,
        type: 'Micro',
        image: 'https://schumacher.tur.br/wp-content/uploads/2022/07/img-9556.jpg',
        features: ['Ar/Calefação', 'Som', 'Tomadas', 'Geladeira elétrica'],
    },
    {
        id: 'sprinter',
        name: 'Sprinter Executiva',
        model: 'Mercedes-Benz Sprinter 415',
        capacity: 18,
        type: 'Van',
        image: 'https://schumacher.tur.br/wp-content/uploads/2022/07/whatsapp-image-2020-07-27-at-15-27-40-800x600.jpg',
        features: ['Poltronas super soft', 'TV/DVD', 'Tomadas', 'Ar-condicionado'],
    },
]

export const allVehicles = [...buses, ...micros]

export const fleetStats = {
    totalVehicles: allVehicles.length,
    totalCapacity: allVehicles.reduce((sum, v) => sum + v.capacity, 0),
    busesCount: buses.length,
    microsCount: micros.length,
}

export function getVehicleById(id) {
    return allVehicles.find(v => v.id === id)
}

export function getVehiclesByType(type) {
    return allVehicles.filter(v => v.type.toLowerCase() === type.toLowerCase())
}
