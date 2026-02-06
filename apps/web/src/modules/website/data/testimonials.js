// Depoimentos reais de clientes
export const testimonials = [
    {
        id: 1,
        name: 'Darlene Xavier',
        destination: 'Lençóis Maranhenses',
        date: '2024',
        rating: 5,
        avatar: '👩',
        text: 'Gostaria de agradecer por tudo, principalmente pela paciência. A viagem foi perfeita… Receptivo maravilhoso, motorista maravilhoso, guia maravilhoso, funcionários e serviços perfeitos. Sem contar o lugar visitado que é simplesmente magnífico e mágico.',
        tags: ['maranhao', 'atendimento', 'organização'],
    },
    {
        id: 2,
        name: 'Carlos Alberto',
        destination: 'Lençóis Maranhenses',
        date: '2024',
        rating: 5,
        avatar: '👨',
        text: 'A experiência no Maranhão foi fantástica e completamente impactante. Os lugares são incríveis e deslumbrantes. Vale lembrar que é um roteiro simples mas magnífico, cheio de natureza e interação com o ambiente! As pessoas são acolhedoras e fazem você se sentir em casa, literalmente.',
        tags: ['maranhao', 'experiência', 'natureza'],
    },
    {
        id: 3,
        name: 'Família Campos',
        destination: 'Viagem em Família',
        date: '2024',
        rating: 5,
        avatar: '👨‍👩‍👧',
        text: 'A viagem foi maravilhosa. Além das belezas naturais, tiro o chapéu para a organização. Tudo feito com muito profissionalismo, pensando nos mínimos detalhes. Super recomendo.',
        tags: ['família', 'organização', 'profissionalismo'],
    },
]

export function getTestimonialsByDestination(destination) {
    return testimonials.filter(t =>
        t.destination.toLowerCase().includes(destination.toLowerCase()) ||
        t.tags.includes(destination.toLowerCase())
    )
}

export function getTestimonialsByTag(tag) {
    return testimonials.filter(t => t.tags.includes(tag.toLowerCase()))
}

export function getAllTestimonials() {
    return testimonials
}
