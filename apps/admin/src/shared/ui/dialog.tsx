import { Modal } from "@heroui/react";
import {
  cloneElement,
  createContext,
  isValidElement,
  useContext,
  type HTMLAttributes,
  type ReactElement,
  type ReactNode,
} from "react";

import { cn } from "./style";

type DialogContextValue = {
  onOpenChange?: (open: boolean) => void;
};

const DialogContext = createContext<DialogContextValue>({});

function DialogRoot({
  children,
  onOpenChange,
  open,
}: {
  children: ReactNode;
  onOpenChange?: (open: boolean) => void;
  open?: boolean;
}) {
  return (
    <DialogContext.Provider value={{ onOpenChange }}>
      <Modal isOpen={open} onOpenChange={onOpenChange}>
        {children}
      </Modal>
    </DialogContext.Provider>
  );
}

function DialogTrigger({ asChild, children }: { asChild?: boolean; children: ReactNode }) {
  const { onOpenChange } = useContext(DialogContext);

  if (asChild && isValidElement(children)) {
    const child = children as ReactElement<{ onClick?: () => void; onPress?: () => void }>;

    return cloneElement(child, {
      onClick: () => {
        child.props.onClick?.();
        onOpenChange?.(true);
      },
      onPress: () => {
        child.props.onPress?.();
        onOpenChange?.(true);
      },
    });
  }

  return <Modal.Trigger>{children}</Modal.Trigger>;
}

function DialogContent({ className, ...props }: HTMLAttributes<HTMLDivElement>) {
  return (
    <Modal.Backdrop>
      <Modal.Container className={cn("w-[min(560px,calc(100vw-32px))]", className)}>
        <Modal.Dialog>
          <Modal.CloseTrigger />
          <div {...props} />
        </Modal.Dialog>
      </Modal.Container>
    </Modal.Backdrop>
  );
}

function DialogHeader({ className, ...props }: HTMLAttributes<HTMLDivElement>) {
  return <Modal.Header className={cn("flex flex-col gap-1.5", className)} {...props} />;
}

function DialogFooter({ className, ...props }: HTMLAttributes<HTMLDivElement>) {
  return (
    <Modal.Footer
      className={cn("flex items-center justify-end gap-2 pt-4", className)}
      {...props}
    />
  );
}

function DialogTitle({ className, ...props }: HTMLAttributes<HTMLHeadingElement>) {
  return (
    <Modal.Heading
      className={cn("text-[16px] leading-[24px] font-semibold", className)}
      {...props}
    />
  );
}

function DialogDescription({ className, ...props }: HTMLAttributes<HTMLParagraphElement>) {
  return (
    <p className={cn("text-[13px] leading-[20px] text-text-secondary", className)} {...props} />
  );
}

export const Dialog = {
  Root: DialogRoot,
  Trigger: DialogTrigger,
  Content: DialogContent,
  Header: DialogHeader,
  Footer: DialogFooter,
  Title: DialogTitle,
  Description: DialogDescription,
};
