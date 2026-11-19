import { useState } from "react";
import type { PublicCustomerService } from "@/lib/customer-service";
import "./customer-service.css";

export function CustomerServiceContent({ value }: { value: PublicCustomerService }) {
    const contacts = [
        { name: value.contactName, phone: value.phone, wechat: value.wechat, qrCodeUrl: value.qrCodeUrl },
        { name: value.secondContactName, phone: value.secondPhone, wechat: value.secondWechat, qrCodeUrl: value.secondQrCodeUrl },
    ].filter((contact) => contact.name || contact.phone || contact.wechat || contact.qrCodeUrl);
    return (
        <div className="customer-service-content">
            {value.description && <p className="customer-service-description">{value.description}</p>}
            <div className="customer-service-contacts">
                {contacts.map((contact, index) => (
                    <CustomerServiceContact key={index} {...contact} />
                ))}
            </div>
        </div>
    );
}

function CustomerServiceContact({ name, phone, wechat, qrCodeUrl }: { name: string; phone: string; wechat: string; qrCodeUrl: string }) {
    const [failedURL, setFailedURL] = useState("");
    return (
        <section className="customer-service-person" aria-label={name || "客服联系方式"}>
            {qrCodeUrl &&
                (failedURL === qrCodeUrl ? (
                    <p role="status" className="text-muted-foreground">
                        二维码暂时无法加载，请使用下方联系方式
                    </p>
                ) : (
                    <img className="customer-service-qr" src={qrCodeUrl} alt={name ? `${name}的客服二维码` : "客服二维码"} onError={() => setFailedURL(qrCodeUrl)} />
                ))}
            {(name || phone || wechat) && (
                <div className="customer-service-contact">
                    {name && <p>{name}</p>}
                    {phone && <p>手机号：{phone}</p>}
                    {wechat && <p>微信号：{wechat}</p>}
                </div>
            )}
        </section>
    );
}
