import { Col, Row } from 'react-bootstrap';
import { useQuery } from '@tanstack/react-query';
import { useNavigate } from 'react-router-dom';
import { CheckFetchStatus, RouteOnNotOK } from '@lib/Status';
import CouponItemShow from '@components/CouponItemShow';

interface ICoupon {
  description: string;
  discount: number;
  expire_date: string;
  id: number;
  name: string;
  scope: 'global' | 'shop';
  start_date: string;
  type: 'percentage' | 'fixed' | 'shipping';
}

const SellerCoupons = () => {
  const navigate = useNavigate();

  // TODO
  // const {sellerName} = useParams();
  const sellerName = 'user1';

  const { data: CouponsData, status: fetchCouponsStatus } = useQuery({
    queryKey: ['GetShopCoupons'],
    queryFn: async () => {
      const resp = await fetch(`/api/shop/${sellerName}/coupon?offset=0&limit=10`, {
        method: 'GET',
        headers: {
          accept: 'application/json',
        },
      });
      if (!resp.ok) {
        RouteOnNotOK(resp, navigate);
      } else {
        return await resp.json();
      }
    },
    select: (data) => data as ICoupon[],
    enabled: true,
    refetchOnWindowFocus: false,
  });

  if (fetchCouponsStatus !== 'success') {
    return <CheckFetchStatus status={fetchCouponsStatus} />;
  }

  const globalCoupons = CouponsData.filter((coupon) => coupon.scope === 'global');
  const shopCoupons = CouponsData.filter((coupon) => coupon.scope === 'shop');

  return (
    <div>
      <Row>
        <Col md={12}>
          <div className='title'>Global Coupon</div>
        </Col>
        <hr className='hr' />
        {globalCoupons.map((data, index) => {
          return (
            <Col xs={12} md={4} xl={3} key={index} style={{ padding: '2%' }}>
              <CouponItemShow
                data={{
                  discount: data.discount,
                  expire_date: data.expire_date,
                  id: data.id,
                  name: data.name,
                  scope: data.scope,
                  type: data.type,
                }}
              />
            </Col>
          );
        })}
        <Col md={12} style={{ paddingTop: '5%' }}>
          <div className='title'>Shop Coupon</div>
        </Col>
        <hr className='hr' />
        {shopCoupons.map((data, index) => {
          return (
            <Col xs={12} md={4} xl={3} key={index} style={{ padding: '2%' }}>
              <CouponItemShow
                data={{
                  discount: data.discount,
                  expire_date: data.expire_date,
                  id: data.id,
                  name: data.name,
                  scope: data.scope,
                  type: data.type,
                }}
              />
            </Col>
          );
        })}
      </Row>
    </div>
  );
};

export default SellerCoupons;
