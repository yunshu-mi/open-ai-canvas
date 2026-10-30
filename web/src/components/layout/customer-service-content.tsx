import { useState } from "react";
import type { PublicCustomerService } from "@/lib/customer-service";
import "./customer-service.css";

export function CustomerServiceContent({ value }: { value: PublicCustomerService }) {
    const [failedURL, setFailedURL] = useState("");
    return (
        <div className="customer-service-content">
            {value.description && <p className="customer-service-description">{value.description}</p>}
            {value.qrCodeUrl &&
                (failedURL === value.qrCodeUrl ? (
                    <p role="status" className="text-muted-foreground">
                        二维码暂时无法加载，请使用下方联系方式
                    </p>
                ) : (
                    <img className="customer-service-qr" src={value.qrCodeUrl} alt="客服二维码" onError={() => setFailedURL(value.qrCodeUrl)} />
                ))}
            {(value.contactName || value.phone || value.wechat) && (
                <div className="customer-service-contact">
                    {value.contactName && <p>{value.contactName}</p>}
                    {value.phone && <p>手机号：{value.phone}</p>}
                    {value.wechat && <p>微信号：{value.wechat}</p>}
                </div>
            )}
        </div>
    );
}
