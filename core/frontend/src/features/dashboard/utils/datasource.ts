export const isPrometheusDatasource = (name: string, type?: string): boolean =>
  type
    ? type.toLowerCase() === 'prometheus'
    : name.toLowerCase().includes('prometheus');
