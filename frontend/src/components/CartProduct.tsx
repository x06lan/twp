import { Col, Row } from 'react-bootstrap';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { faTrash } from '@fortawesome/free-solid-svg-icons';

interface Props {
  data: ProductProps;
  cart_id: number;
  onRefetch: () => void;
}

interface ProductProps {
  enabled: boolean;
  image_id: string;
  name: string;
  price: number;
  product_id: number;
  quantity: number;
  stock: number;
}

const CartProduct = ({ data, cart_id, onRefetch }: Props) => {
  // TODO: Buyer delete product in cart
  // DELETE /buyer/cart/:cart_id/product/:product_id
  const removeItem = () => {
    console.log(`${data.name} delete in cart ${cart_id}`);
    onRefetch();
  };

  const updateQuantity = (quantity: number) => {
    // TODO: Buyer edit product in cart
    // PATCH /buyer/cart/:cart_id/product/:product_id
    // body: { quantity: number }
    if (quantity === 0) {
      removeItem();
    } else if (quantity > 0 && quantity <= data.stock) {
      console.log(`${data.name} ${quantity} in cart ${cart_id}`);
      onRefetch();
    }
  };

  return (
    <div className='cart_item' style={{ margin: '2% 0 2% 0' }}>
      <Row>
        <Col xs={4} md={1} className='center'>
          <img src={data.image_id} style={{ width: '100%', borderRadius: '10px' }} />
        </Col>
        <Col xs={8} md={11} className='dark center_vertical'>
          <div className='disappear_phone' style={{ width: '100%' }}>
            <Row className='center_vertical' style={{ width: '100%' }}>
              <Col
                md={4}
                className='center_vertical'
                style={{ wordBreak: 'break-all', fontSize: '20px' }}
              >
                {data.name}
              </Col>
              <Col md={5} className='center' style={{ padding: '2% 0' }}>
                <Row style={{ padding: '0', margin: '0' }}>
                  <Col md={3} onClick={() => updateQuantity(data.quantity - 1)} className='pointer'>
                    <div className='quantity_f pointer center '>-</div>
                  </Col>
                  <Col md={6} className='center'>
                    <div>
                      <input
                        type='text'
                        className='quantity_box'
                        value={data.quantity}
                        onChange={(e) => updateQuantity(parseInt(e.target.value))}
                        style={{ textAlign: 'center' }}
                      />
                    </div>
                  </Col>
                  <Col md={3} onClick={() => updateQuantity(data.quantity + 1)} className='pointer'>
                    <div className='quantity_f pointer center'>+</div>
                  </Col>
                </Row>
              </Col>
              <Col md={2} className='right ' style={{ padding: '2% 0', fontSize: '20px' }}>
                {data.price * data.quantity} NTD
              </Col>
              <Col md={1} className='center' style={{ padding: '2% 0' }}>
                <FontAwesomeIcon icon={faTrash} size='xl' className='trash' onClick={removeItem} />
              </Col>
            </Row>
          </div>

          <div className='disappear_tablet disappear_desktop'>
            <Row className='center_vertical' style={{ width: '100%' }}>
              <Col
                xs={12}
                className='center_vertical'
                style={{ wordBreak: 'break-all', padding: '2% 0 0 0' }}
              >
                {data.name}
              </Col>
              <Col xs={12} className='center' style={{ padding: '2% 0 0 0' }}>
                <Row style={{ padding: '0', margin: '0', width: '100%' }}>
                  <Col
                    xs={3}
                    onClick={() => updateQuantity(data.quantity - 1)}
                    className='pointer'
                    style={{ padding: '0 2%', margin: '0' }}
                  >
                    <div className='quantity_f pointer center '>-</div>
                  </Col>
                  <Col xs={6} className='center' style={{ padding: '0 2%', margin: '0' }}>
                    <div>
                      <input
                        type='text'
                        className='quantity_box'
                        value={data.quantity}
                        onChange={(e) => updateQuantity(parseInt(e.target.value))}
                        style={{ textAlign: 'center' }}
                      />
                    </div>
                  </Col>
                  <Col
                    xs={3}
                    onClick={() => updateQuantity(data.quantity + 1)}
                    className='pointer'
                    style={{ padding: '0 2%', margin: '0' }}
                  >
                    <div className='quantity_f pointer center'>+</div>
                  </Col>
                </Row>
              </Col>
              <Col xs={8} style={{ padding: '2% 0 2% 5%' }}>
                {data.price * data.quantity} NTD
              </Col>

              <Col xs={4} className='right' style={{ padding: '2% 5% 2% 0' }}>
                <FontAwesomeIcon icon={faTrash} className='trash' onClick={removeItem} />
              </Col>
            </Row>
          </div>
        </Col>
      </Row>
    </div>
  );
};

export default CartProduct;
