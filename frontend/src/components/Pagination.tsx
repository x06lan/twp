import { useForm } from 'react-hook-form';

interface Props {
  searchParams: URLSearchParams;
  setSearchParams: (params: URLSearchParams, options?: { replace?: boolean }) => void;
  refresh: () => void;
  limit?: number;
  maxPage?: number;
  isMore?: boolean;
}

const Pagination = ({
  searchParams,
  setSearchParams,
  refresh,
  limit = 10,
  maxPage = 100,
  isMore = true,
}: Props) => {
  if (!searchParams.has('offset')) {
    searchParams.set('offset', '0');
  }
  if (!searchParams.has('limit') || Number(searchParams.get('limit')) !== limit) {
    searchParams.set('limit', limit.toString());
  }
  const getPage = () => {
    return Number(searchParams.get('offset')) / limit + 1;
  };

  const { register, handleSubmit, setValue } = useForm<{ newPage: number }>({
    defaultValues: { newPage: getPage() },
  });

  const onPrevious = () => {
    const page = getPage();
    if (page > 1) {
      searchParams.set('offset', ((page - 2) * limit).toString());
      setSearchParams(searchParams, { replace: true });
      refresh();
    }
  };
  const onNext = () => {
    const page = getPage();
    if (page < maxPage && isMore) {
      searchParams.set('offset', (page * limit).toString());
      setSearchParams(searchParams, { replace: true });
      refresh();
    }
  };

  const onSubmit = (data: { newPage: number }) => {
    const inputPage = data.newPage;
    if (inputPage > 0 && inputPage < maxPage && (isMore || (!isMore && inputPage < getPage()))) {
      searchParams.set('offset', ((inputPage - 1) * limit).toString());
      setSearchParams(searchParams, { replace: true });
      refresh();
    } else {
      setValue('newPage', getPage());
    }
  };
  return (
    <div className='pagination center_vertical center'>
      <div className='center' onClick={onPrevious}>
        {'<'}
      </div>
      <div className='center'>{'Page: '}</div>
      <form onSubmit={handleSubmit(onSubmit)}>
        <input
          className='center'
          {...register('newPage')}
          style={{
            maxWidth: '50px',
            border: 'none',
            textAlign: 'center',
            borderRadius: '10px',
            fontWeight: 'bold',
            color: 'var(--button_light)',
          }}
        />
      </form>
      <div className='center' onClick={onNext}>
        {'>'}
      </div>
    </div>
  );
};

export default Pagination;
