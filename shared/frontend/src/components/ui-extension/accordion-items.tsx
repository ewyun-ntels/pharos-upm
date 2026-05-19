import React from 'react';
import {
  Accordion as CustomAccordion,
  AccordionItem,
  AccordionTrigger,
  AccordionContent,
} from './accordion-custom';

interface AccordionProps {
  items: {
    id: string;
    title: string;
    content: React.ReactNode;
  }[];
  defaultOpenItems?: string[] | 'all';
  type?: 'single' | 'multiple';
  id?: string;
}

/**
 * Accordion with items array API
 * 
 * Provides a convenient items-based API for Accordion component
 * that simplifies usage in panel options and other repetitive scenarios.
 */
const Accordion: React.FC<AccordionProps> = ({
  id,
  items,
  defaultOpenItems = [],
  type = 'single',
}) => {
  const defaultValue: string[] =
    defaultOpenItems === 'all' ? items.map((item) => item.id) : defaultOpenItems;

  if (type === 'multiple') {
    return (
      <CustomAccordion key={'accordion-' + id} type="multiple" defaultValue={defaultValue}>
        {items.map((item) => (
          <AccordionItem key={item.id} value={item.id}>
            <AccordionTrigger>{item.title}</AccordionTrigger>
            <AccordionContent>{item.content}</AccordionContent>
          </AccordionItem>
        ))}
      </CustomAccordion>
    );
  } else if (type === 'single') {
    return (
      <CustomAccordion
        key={'accordion-' + id}
        type="single"
        collapsible
        defaultValue={defaultValue[0]}
      >
        {items.map((item) => (
          <AccordionItem key={item.id} value={item.id}>
            <AccordionTrigger>{item.title}</AccordionTrigger>
            <AccordionContent>{item.content}</AccordionContent>
          </AccordionItem>
        ))}
      </CustomAccordion>
    );
  }
};

export default Accordion;
