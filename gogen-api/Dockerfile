FROM public.ecr.aws/lambda/python:3.13

# Copy requirements.txt
COPY requirements.txt ${LAMBDA_TASK_ROOT}

# Install the specified packages
RUN pip install -r requirements.txt -t ${LAMBDA_TASK_ROOT}

# Copy function code
COPY api/ ${LAMBDA_TASK_ROOT}/

# Set the CMD to your handler
CMD [ "get.lambda_handler" ] 